package neo4j

var Constraints = []string{
	"CREATE CONSTRAINT stock_ticker IF NOT EXISTS FOR (s:Stock) REQUIRE s.ticker IS UNIQUE",
	"CREATE CONSTRAINT event_id IF NOT EXISTS FOR (e:Event) REQUIRE e.id IS UNIQUE",
	"CREATE INDEX company_name IF NOT EXISTS FOR (c:Company) ON (c.name)",
	"CREATE INDEX event_date IF NOT EXISTS FOR (e:Event) ON (e.date)",
}

func UpsertCompanyCypher() string {
	return `MERGE (c:Company {name: $name})
SET c.also_known_as = $also_known_as
RETURN c.name AS name`
}

func UpsertStockCypher() string {
	return `MERGE (s:Stock {ticker: $ticker})
SET s.former_tickers = $former_tickers,
    s.status = $status
WITH s
MATCH (c:Company {name: $company_name})
MERGE (c)-[:ISSUES]->(s)
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
