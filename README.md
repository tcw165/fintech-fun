# fintech-fun

Playground for fin-tech ideas.

## Build

Bazel 8.2 + Go 1.24 (rules_go) + Python 3.12 for the hello ADK agent. Refresh the pip lock after editing `requirements.in`:

```bash
bazel run //:requirements.update
```

Run the hello Google ADK agent tests (no live Gemini calls):

```bash
bazel test //agents/skills/hello:hello_agent_test
```

## Local cluster (minikube)

Neo4j, Qdrant, `api_server`, and ingest CronJobs run in a single minikube profile. No Docker Compose.

```bash
minikube start --profile=fintech-fun --driver=docker
bazel build //api_server/cmd:api_server //ingest_jobs/corporate_actions:corporate_actions

mkdir -p /tmp/fintech-fun-images/api-server /tmp/fintech-fun-images/corporate-actions
cp "$(bazel cquery --output=files //api_server/cmd:api_server)" /tmp/fintech-fun-images/api-server/api_server
cp "$(bazel cquery --output=files //ingest_jobs/corporate_actions:corporate_actions)" /tmp/fintech-fun-images/corporate-actions/corporate_actions
docker build -f api_server/Dockerfile -t fintech-fun/api-server:local /tmp/fintech-fun-images/api-server
docker build -f ingest_jobs/corporate_actions/Dockerfile -t fintech-fun/corporate-actions:local /tmp/fintech-fun-images/corporate-actions

minikube image load -p fintech-fun fintech-fun/api-server:local
minikube image load -p fintech-fun fintech-fun/corporate-actions:local

kubectl --context fintech-fun apply -k infra/k8s/overlays/local
minikube service -p fintech-fun -n fintech-fun api-server --url
k9s --context fintech-fun
```

| Service | In-cluster | NodePort (local overlay) | Auth |
| --- | --- | --- | --- |
| Neo4j Browser | `neo4j:7474` | `:30474` | `neo4j` / `fintechfun` |
| Neo4j Bolt | `neo4j:7687` | `:30687` | `neo4j` / `fintechfun` |
| Qdrant HTTP | `qdrant:6333` | `:30333` | none (local) |
| Qdrant gRPC | `qdrant:6334` | `:30334` | none (local) |
| api_server | `api-server:8080` | `:30080` | none (local) |

The corporate-actions CronJob is owned by `infra/` and starts **suspended** until a live `run` subcommand exists.

Stop the cluster: `minikube stop -p fintech-fun`. Delete it: `minikube delete -p fintech-fun`.

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
//agents/harness/contract                             Skill, Runner interfaces
//agents/harness/impl                                   default Runner → Skill.Run
//agents/skills/ingest/robinhood/corporate_actions      models: hood_events + parser/classify/ingest/skill
//agents/skills/hello                                   Python ADK demo skill
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

The ingest job reads [Robinhood Corporate Actions Tracker](https://robinhood.com/us/en/support/articles/corporate-actions-tracker/) text day by day, classifies each row into `Event.kind`, and previews Neo4j/Qdrant writes:

```bash
bazel run //ingest_jobs/corporate_actions -- parse "$PWD/agents/skills/ingest/robinhood/corporate_actions/testdata/tracker_sept_2026.txt"
bazel run //ingest_jobs/corporate_actions -- classify 2026-09-03 'Apogee Therapeutics, Inc. (APGE) performed a cash merger.'
bazel run //ingest_jobs/corporate_actions -- plan "$PWD/agents/skills/ingest/robinhood/corporate_actions/testdata/tracker_sept_2026.txt"
```
