# fintech-fun

Playground for fin-tech ideas.

## Build

Bazel 8.2 + Go 1.24 (rules_go) + Python 3.12 for the existing hello ADK agent. Refresh the pip lock after editing `requirements.in`:

```bash
bazel run //:requirements.update
```

Run the hello Google ADK agent tests (no live Gemini calls):

```bash
bazel test //agents/hello:hello_agent_test
```

## Local databases

Neo4j (graph) and Qdrant (vectors) run via Docker Compose. Data lives in named volumes.

```bash
docker compose -f infra/docker-compose.yml up -d
docker compose -f infra/docker-compose.yml ps
```

| Service | URL | Auth |
| --- | --- | --- |
| Neo4j Browser | http://localhost:7474 | `neo4j` / `fintechfun` |
| Neo4j Bolt | bolt://localhost:7687 | `neo4j` / `fintechfun` |
| Qdrant HTTP | http://localhost:6333 | none (local) |
| Qdrant gRPC | localhost:6334 | none (local) |

Stop without deleting data: `docker compose -f infra/docker-compose.yml down`. Wipe volumes: add `-v`.

## Go layout

- Models and contracts are their own Bazel modules (pure data or protocol).
- Client interfaces and implementations live in different modules.
- Dependencies point child → parent. Parents never import children.

```
//graph                                                 models: Company, Stock, Event
//graph/examples                                        fixtures → graph
//graph/fold                                            query → graph
//graph/clients/neo4j                                   contract: Client + Cypher (no driver)
//graph/clients/neo4j/impl                              impl → neo4j
//graph/clients/qdrant                                  contract: Client + collection protocol (no HTTP)
//graph/clients/qdrant/impl                             impl → qdrant
//ingest/agents/robinhood/corporate_actions             models: hood_events (TrackerRow, ClassifiedEvent)
//ingest/agents/robinhood/corporate_actions/testdata    fixture → models
//ingest/agents/robinhood/corporate_actions/parser      child → models
//ingest/agents/robinhood/corporate_actions/classify    child → models
//ingest/agents/robinhood/corporate_actions/ingest      child → models + graph contracts
//ingest/agents/robinhood/corporate_actions/crawler     CLI → parser, classify, ingest
```

Retail graph tests (no live Docker):

```bash
bazel test //graph:schema_test //graph/fold:fold_test //graph/clients/neo4j:neo4j_test //graph/clients/qdrant:qdrant_test
bazel test //ingest/agents/robinhood/corporate_actions/parser:parser_test //ingest/agents/robinhood/corporate_actions/classify:classify_test //ingest/agents/robinhood/corporate_actions/ingest:ingest_test
```

The crawler reads [Robinhood Corporate Actions Tracker](https://robinhood.com/us/en/support/articles/corporate-actions-tracker/) text day by day, classifies each row into `Event.kind`, and previews Neo4j/Qdrant writes:

```bash
bazel run //ingest/agents/robinhood/corporate_actions/crawler -- parse "$PWD/ingest/agents/robinhood/corporate_actions/testdata/tracker_sept_2026.txt"
bazel run //ingest/agents/robinhood/corporate_actions/crawler -- classify 2026-09-03 'Apogee Therapeutics, Inc. (APGE) performed a cash merger.'
bazel run //ingest/agents/robinhood/corporate_actions/crawler -- plan "$PWD/ingest/agents/robinhood/corporate_actions/testdata/tracker_sept_2026.txt"
```
