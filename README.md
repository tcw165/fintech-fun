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
just search "LivePerson stock merger"  # GET /search?q=… through minikube
just k9s      # attach to context fintech-fun
just down     # stop the VM
```

`just` lists every recipe. `just cluster-start` / `cluster-images` / `cluster-apply` / `just wait` are the steps inside `up` if you want them one at a time.

## Gap A (local minikube proof)

Activation product code is merged. This checklist proves it on a live local cluster:

```bash
just deps
just up          # start profile, load images, apply overlay, wait for Ready
just healthz
just ping        # Bolt + Qdrant on NodePorts :30687 / :30333
just seed        # Notion fixtures (gold path)
just verify      # memory gold + live Bolt gold (nflx…ftel)
just smoke       # /fold /search /v1/graph
just gold        # HTTP fold rows match the Notion tables
just gap-a       # wait + healthz + ping + seed + verify + smoke + gold
```

The corporate-actions CronJob stays **suspended**. `just refresh` is the manual refresh path. Unsuspend only after Robinhood is reachable:

```bash
just egress && just unsuspend
```

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
- `di/` is for runnable modules only (servers, jobs, binaries). Library modules take injected interfaces.

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
//ingest_jobs/di                                        composition root: AppContext + clients
//ingest_jobs/corporate_actions                         thin job binary → harness + ingest skill
//api_server/contract                                   HealthHandler interface
//api_server/impl                                       HTTP /healthz
//api_server/di                                         composition root: AppContext + clients
//api_server/cmd                                        process entrypoint
//infra/k8s                                             minikube manifests + CronJob
```

Sync `BUILD.bazel` deps from Go imports: `just update-build-files` (`bazel run //:gazelle`).

Retail graph tests (no live cluster):

```bash
bazel test //graph:schema_test //graph/fold:fold_test //graph/clients/neo4j:neo4j_test //graph/clients/qdrant:qdrant_test
bazel test //agents/harness/impl:impl_test
bazel test //agents/skills/ingest/robinhood/corporate_actions/parser:parser_test //agents/skills/ingest/robinhood/corporate_actions/classify:classify_test //agents/skills/ingest/robinhood/corporate_actions/ingest:ingest_test
bazel test //api_server/impl:impl_test //api_server/di:di_test
bazel test //ingest_jobs/di:di_test
```

`GET /fold?q=square&qty=10` is the account-screen query. `GET /search?q=` ranks `Event.headline` in Qdrant. `api_server` injects Bolt and Qdrant at process start (`NEO4J_URI` / `QDRANT_URL`; in-cluster `bolt://neo4j:7687` / `http://qdrant:6333`).

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
