package neo4j

var Constraints = []string{
	"CREATE CONSTRAINT stock_ticker IF NOT EXISTS FOR (s:Stock) REQUIRE s.ticker IS UNIQUE",
	"CREATE CONSTRAINT event_id IF NOT EXISTS FOR (e:Event) REQUIRE e.id IS UNIQUE",
	"CREATE CONSTRAINT ingest_source_id IF NOT EXISTS FOR (s:IngestSource) REQUIRE s.id IS UNIQUE",
	"CREATE INDEX company_name IF NOT EXISTS FOR (c:Company) ON (c.name)",
	"CREATE INDEX event_date IF NOT EXISTS FOR (e:Event) ON (e.date)",
}

func UpsertCompanyCypher() string {
	return UpsertIssuerCypher()
}

func UpsertStockCypher() string {
	return UpsertIssuerCypher()
}

// UpsertIssuerCypher attaches Company through Stock. Name is a display field, not a MERGE key.
func UpsertIssuerCypher() string {
	return `MERGE (s:Stock {ticker: $ticker})
SET s.former_tickers = $former_tickers,
    s.status = $status
WITH s
OPTIONAL MATCH (c:Company)-[:ISSUES]->(s)
WITH s, c
FOREACH (_ IN CASE WHEN c IS NULL THEN [1] ELSE [] END |
  CREATE (n:Company {name: $company_name, also_known_as: $also_known_as})-[:ISSUES]->(s)
)
FOREACH (_ IN CASE WHEN c IS NOT NULL THEN [1] ELSE [] END |
  SET c.name = $company_name,
      c.also_known_as = $also_known_as
)
RETURN s.ticker AS ticker`
}

// FoldCypher is the account-screen query from the Notion retail graph page.
func FoldCypher() string {
	return `WITH toLower($q) AS q, toFloat($qty) AS qty0

MATCH (c:Company)-[:ISSUES]->(s:Stock)
WHERE toLower(c.name) CONTAINS q
   OR any(a IN coalesce(c.also_known_as, []) WHERE toLower(a) CONTAINS q)
   OR toLower(s.ticker) = q
   OR any(t IN coalesce(s.former_tickers, []) WHERE toLower(t) = q)

MATCH (e:Event)-[:HAPPENED_TO]->(s)
OPTIONAL MATCH (e)-[:YOU_NOW_HOLD]->(after:Stock)

WITH c, s, qty0, e, after
ORDER BY e.date

WITH c, s, qty0, collect({
  date: e.date,
  kind: e.kind,
  headline: e.headline,
  share_multiplier: e.share_multiplier,
  cash_per_share: coalesce(e.cash_per_share, 0.0),
  now_holds: after.ticker
}) AS events

WITH c, s, qty0,
     reduce(
       acc = {qty: qty0, cash: 0.0, rows: []},
       ev IN events |
         {
           qty:  CASE ev.kind WHEN 'reverse_split' THEN acc.qty / ev.share_multiplier ELSE acc.qty * ev.share_multiplier END,
           cash: acc.cash + acc.qty * ev.cash_per_share,
           rows: acc.rows + [{
             date:            ev.date,
             kind:            ev.kind,
             qty_before:      acc.qty,
             qty_after:       CASE ev.kind WHEN 'reverse_split' THEN acc.qty / ev.share_multiplier ELSE acc.qty * ev.share_multiplier END,
             cash_this_event: acc.qty * ev.cash_per_share,
             now_holds:       ev.now_holds
           }]
         }
     ) AS fold

RETURN
  c.name     AS company,
  s.ticker   AS ticker_now,
  s.status   AS status,
  qty0       AS qty_started,
  fold.qty   AS qty_now,
  fold.cash  AS cash_received,
  fold.rows  AS series`
}

func Fold(client Client, q string, qty float64) ([]map[string]any, error) {
	return client.Run(FoldCypher(), map[string]any{"q": q, "qty": qty})
}

// SeriesCypher resolves Company + Stock + Event series without folding qty.
func SeriesCypher() string {
	return `WITH toLower($q) AS q

MATCH (c:Company)-[:ISSUES]->(s:Stock)
WHERE toLower(c.name) CONTAINS q
   OR any(a IN coalesce(c.also_known_as, []) WHERE toLower(a) CONTAINS q)
   OR toLower(s.ticker) = q
   OR any(t IN coalesce(s.former_tickers, []) WHERE toLower(t) = q)

MATCH (e:Event)-[:HAPPENED_TO]->(s)
OPTIONAL MATCH (e)-[:YOU_NOW_HOLD]->(after:Stock)

WITH c, s, e, after
ORDER BY e.date

RETURN
  c.name   AS company,
  s.ticker AS ticker_now,
  s.status AS status,
  collect({
    id: e.id,
    date: e.date,
    kind: e.kind,
    headline: e.headline,
    share_multiplier: e.share_multiplier,
    cash_per_share: coalesce(e.cash_per_share, 0.0),
    now_holds: after.ticker
  }) AS events`
}

func Series(client Client, q string) ([]map[string]any, error) {
	return client.Run(SeriesCypher(), map[string]any{"q": q})
}

func ListEventIDsCypher() string {
	return `MATCH (e:Event) RETURN e.id AS id`
}

func ReadSourceCypher() string {
	return `MATCH (s:IngestSource {id: $id}) RETURN s.page_sha256 AS page_sha256, s.fetched_at AS fetched_at, s.history_prefix AS history_prefix`
}

func UpsertSourceCypher() string {
	return `MERGE (s:IngestSource {id: $id})
SET s.page_sha256 = $page_sha256,
    s.fetched_at = datetime()
RETURN s.id AS id, s.page_sha256 AS page_sha256`
}

func UpsertSourceStateCypher() string {
	return `MERGE (s:IngestSource {id: $id})
SET s.page_sha256 = $page_sha256,
    s.history_prefix = $history_prefix,
    s.fetched_at = datetime()
RETURN s.id AS id, s.page_sha256 AS page_sha256, s.history_prefix AS history_prefix`
}

func UpsertEventCypher() string {
	return `MERGE (e:Event {id: $id})
ON CREATE SET
  e.date = date($date),
  e.kind = $kind,
  e.headline = $headline,
  e.share_multiplier = $share_multiplier,
  e.cash_per_share = $cash_per_share,
  e.keep_fractionals = $keep_fractionals,
  e.can_trade = $can_trade
WITH e
MATCH (s:Stock {ticker: $happened_to})
MERGE (e)-[:HAPPENED_TO]->(s)
FOREACH (_ IN CASE WHEN $you_now_hold IS NULL THEN [] ELSE [1] END |
  MERGE (after:Stock {ticker: $you_now_hold})
  MERGE (e)-[:YOU_NOW_HOLD]->(after)
)
RETURN e.id AS id, e.kind AS kind`
}
