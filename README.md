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

`//graph` is the first model module: Company, Stock, Event. No store clients.

