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
