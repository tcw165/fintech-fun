# fintech-fun

Playground for fin-tech ideas.

## Build

Bazel 8.2 + Go 1.24 (rules_go).

## Local cluster (minikube)

Neo4j, Qdrant, `api_server`, and ingest CronJobs run in one minikube profile (`fintech-fun`) with `--driver=docker`. On a Mac host, **Docker Desktop must be installed and running** — `just up` / `just down` talk to that daemon. Also need `minikube`, `kubectl`, `just`, Bazel, and optionally `k9s`. `just deps` checks Docker.

```bash
just up       # start profile, load images, apply infra/k8s/overlays/local
just healthz  # GET /healthz through minikube
just fold square 10  # GET /fold?q=square&qty=10 through minikube
just k9s      # attach to context fintech-fun
just down     # stop the VM
```

`just` lists every recipe. `just cluster-start` / `cluster-images` / `cluster-apply` are the three steps inside `up` if you want them one at a time.

| Service | In-cluster | NodePort (local overlay) | Auth |
| --- | --- | --- | --- |
| Neo4j Browser | `neo4j:7474` | `:30474` | `neo4j` / `fintechfun` |
| Neo4j Bolt | `neo4j:7687` | `:30687` | `neo4j` / `fintechfun` |
| Qdrant HTTP | `qdrant:6333` | `:30333` | none (local) |
| Qdrant gRPC | `qdrant:6334` | `:30334` | none (local) |
| api_server | `api-server:8080` | `:30080` | none (local) |

The corporate-actions CronJob is owned by `infra/`, starts **suspended**, and is wired to `refresh`. Unsuspend it after the local cluster can reach Robinhood. Locally: `just refresh` after new tracker days.

Delete the profile: `just cluster-delete`.

## Go layout

- Models and contracts are their own Bazel modules (pure data or protocol).
- Client interfaces and implementations live in different modules.
- Dependencies point child → parent. Parents never import children.

```
//graph                                                 models: Company, Stock, Event
//graph/examples                                        fixtures → graph
//graph/fold                                            query → graph
//graph/clients/neo4j                                   contract: Client + Cypher (no driver)
//graph/clients/neo4j/impl                              official Bolt driver + Recording
//graph/clients/qdrant                                  contract: Client + collection protocol (no HTTP)
//graph/clients/qdrant/impl                             official REST client + Recording
//agents/harness/contract                             Skill, Runner interfaces
//agents/harness/impl                                   default Runner → Skill.Run
//agents/skills/ingest/robinhood/corporate_actions      models: hood_events + parser/classify/ingest/skill
//ingest_jobs/corporate_actions                         thin job binary → harness + ingest skill
//api_server/contract                                   HealthHandler interface
//api_server/impl                                       HTTP /healthz
//api_server/cmd                                        process entrypoint
//infra/k8s                                             minikube manifests + CronJob
```

Retail graph tests (no live cluster):

```bash
bazel test //graph:schema_test //graph/fold:fold_test //graph/clients/neo4j:neo4j_test //graph/clients/qdrant:qdrant_test
bazel test //agents/harness/impl:impl_test
bazel test //agents/skills/ingest/robinhood/corporate_actions/parser:parser_test //agents/skills/ingest/robinhood/corporate_actions/classify:classify_test //agents/skills/ingest/robinhood/corporate_actions/ingest:ingest_test
bazel test //api_server/impl:impl_test
```

`GET /fold?q=square&qty=10` is the account-screen query. `api_server` injects the Bolt client at process start (`NEO4J_URI`, in-cluster `bolt://neo4j:7687`).

The ingest job reads [Robinhood Corporate Actions Tracker](https://robinhood.com/us/en/support/articles/corporate-actions-tracker/) text day by day, classifies each row into `Event.kind`, and previews Neo4j/Qdrant writes. `ping` talks to a live cluster (defaults `bolt://localhost:7687` / `http://localhost:6333`, or `NEO4J_URI` / `QDRANT_URL`):

```bash
bazel run //ingest_jobs/corporate_actions -- parse "$PWD/agents/skills/ingest/robinhood/corporate_actions/testdata/tracker_sept_2026.txt"
bazel run //ingest_jobs/corporate_actions -- classify 2026-09-03 'Apogee Therapeutics, Inc. (APGE) performed a cash merger.'
bazel run //ingest_jobs/corporate_actions -- plan "$PWD/agents/skills/ingest/robinhood/corporate_actions/testdata/tracker_sept_2026.txt"
bazel run //ingest_jobs/corporate_actions -- fetch "$PWD/tracker.txt"
bazel run //ingest_jobs/corporate_actions -- gold "$PWD/tracker.txt"
bazel run //ingest_jobs/corporate_actions -- ingest "$PWD/tracker.txt"
bazel run //ingest_jobs/corporate_actions -- verify
bazel run //ingest_jobs/corporate_actions -- search "LivePerson stock merger"
bazel run //ingest_jobs/corporate_actions -- refresh   # or: just refresh
bazel run //ingest_jobs/corporate_actions -- ping
bazel run //ingest_jobs/corporate_actions -- seed
bazel run //ingest_jobs/corporate_actions -- fold square 10
```
