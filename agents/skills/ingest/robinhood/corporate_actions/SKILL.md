---
name: corporate-actions
description: Ingest Robinhood corporate-actions headlines into the retail knowledge graph (Company, Stock, Event) and fold what a holder has now. Use when fetching the tracker, classifying English, constructing or refreshing Neo4j/Qdrant, or answering holder questions.
license: Apache-2.0
compatibility: Go Bazel job; Neo4j Bolt and Qdrant REST; no Python Agent SDK.
allowed-tools: parse classify plan fetch gold ingest verify search refresh ping seed fold
metadata:
  default-tool: ingest
  source: robinhood/corporate_actions
---

# Corporate actions

Drive this skill through the in-repo Agent SDK. The markdown is the source of
truth for the tool list. Empty args run `ingest`.

## Graph to construct

- **Company** — display name plus `also_known_as`. Name is not a MERGE key.
- **Stock** — ticker is identity. Attach Company with `ISSUES`.
- **Event** — id is `(happened_to ticker)|(date)|(kind)`. `HAPPENED_TO` the
  stock; `YOU_NOW_HOLD` only for `now_different_stock` and `extra_stock`.
- Cash lives on Event (`cash_per_share`), not as a node.

## Workflow

1. `fetch` the live tracker (or read a saved page).
2. `parse` rows, then `classify` each headline into an Event.kind.
3. `ingest` (or `refresh`) MERGEs Company/Stock/Event and upserts headlines
   in Qdrant. Re-runs must skip work already done for the same page or Event id.
4. `verify` against the Notion gold fold tables (memory plus live Bolt when
   clients are injected). `fold` answers the account screen. `search` ranks
   Event.headline. On a local minikube cluster, `just gap-a` runs wait,
   healthz, ping, seed, verify, HTTP smoke, and HTTP gold.

## Waiting policy

A `waiting` Event keeps id `(ticker)|(date)|waiting`. A later `cashed_out` or
`now_different_stock` is a new Event on a later date, not an in-place update.
