#!/usr/bin/env bash
# Per-boot startup for fintech-fun: bring up the Docker daemon and the
# Neo4j + Qdrant containers on the localhost ports the apps default to
# (bolt://localhost:7687, http://localhost:6333), then seed the graph once
# so /fold, /search and /v1/graph return data immediately.
# Must be idempotent and tolerate restarts.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

NEO4J_IMAGE="neo4j:5.26.0-community"
QDRANT_IMAGE="qdrant/qdrant:v1.15.5"
NEO4J_PASSWORD="fintechfun"

log() { printf '\n=== %s ===\n' "$*"; }

log "Docker daemon"
if ! sudo docker info >/dev/null 2>&1; then
  sudo bash -c 'nohup dockerd >/var/log/dockerd.log 2>&1 &'
  for _ in $(seq 1 60); do sudo docker info >/dev/null 2>&1 && break; sleep 1; done
fi
sudo docker info >/dev/null 2>&1 || { echo "docker daemon failed to start" >&2; exit 1; }

# ensure_container <name> <image> <run-args...>
ensure_container() {
  local name="$1" image="$2"; shift 2
  if sudo docker ps --format '{{.Names}}' | grep -qx "$name"; then
    return 0
  fi
  if sudo docker ps -a --format '{{.Names}}' | grep -qx "$name"; then
    sudo docker start "$name" >/dev/null
  else
    sudo docker run -d --name "$name" "$@" "$image" >/dev/null
  fi
}

log "Neo4j 5.26 (7474 browser / 7687 bolt)"
ensure_container neo4j "$NEO4J_IMAGE" \
  -p 7474:7474 -p 7687:7687 -e "NEO4J_AUTH=neo4j/${NEO4J_PASSWORD}"

log "Qdrant v1.15.5 (6333 http / 6334 grpc)"
ensure_container qdrant "$QDRANT_IMAGE" \
  -p 6333:6333 -p 6334:6334

log "Wait for services"
for _ in $(seq 1 60); do
  # Read logs into a var first; piping to `grep -q` under pipefail can
  # SIGPIPE `docker logs` and mask the match.
  logs="$(sudo docker logs neo4j 2>&1 || true)"
  case "$logs" in *"Bolt enabled on"*) break ;; esac
  sleep 1
done
for _ in $(seq 1 60); do
  curl -fsS http://localhost:6333/readyz >/dev/null 2>&1 && break
  sleep 1
done

log "Seed graph once (guarded: only when Neo4j is empty)"
NODE_COUNT="$(sudo docker exec neo4j cypher-shell -u neo4j -p "$NEO4J_PASSWORD" \
  --format plain "MATCH (n) RETURN count(n);" 2>/dev/null | tail -1 | tr -dc '0-9')"
NODE_COUNT="${NODE_COUNT:-0}"
if [ "$NODE_COUNT" = "0" ]; then
  # Clear any partial Qdrant collection so seed's create step does not 409.
  curl -fsS -X DELETE http://localhost:6333/collections/corporate_actions >/dev/null 2>&1 || true
  bazel run //ingest_jobs/corporate_actions -- seed || \
    echo "seed skipped (non-fatal)" >&2
else
  echo "Neo4j already has ${NODE_COUNT} nodes; skipping seed"
fi

log "start.sh complete"
