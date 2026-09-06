#!/usr/bin/env bash
# Idempotent environment bootstrap for fintech-fun (Bazel 8.2 + Go 1.24).
# Runs after checkout. Installs the toolchain, warms the Bazel cache, and
# pre-pulls the Neo4j/Qdrant images so per-boot startup (start.sh) is fast.
# Must terminate and be safe to re-run.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

BAZELISK_VERSION="v1.25.0"
NEO4J_IMAGE="neo4j:5.26.0-community"
QDRANT_IMAGE="qdrant/qdrant:v1.15.5"

log() { printf '\n=== %s ===\n' "$*"; }

log "Bazel (bazelisk ${BAZELISK_VERSION}, honors .bazelversion)"
if ! command -v bazel >/dev/null 2>&1; then
  curl -fsSL -o /tmp/bazelisk \
    "https://github.com/bazelbuild/bazelisk/releases/download/${BAZELISK_VERSION}/bazelisk-linux-amd64"
  sudo install -m 0755 /tmp/bazelisk /usr/local/bin/bazel
  rm -f /tmp/bazelisk
fi
bazel version | head -1

log "Docker engine (used for local Neo4j + Qdrant)"
if ! command -v docker >/dev/null 2>&1; then
  sudo apt-get update -qq
  # docker.io is enough; the fuse-overlayfs postinstall can fail harmlessly
  # because we use the vfs storage driver (nested VM cannot convert whiteouts).
  sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq docker.io || true
fi
docker --version

log "Docker daemon config (vfs storage driver)"
# Native/rootless overlayfs cannot convert whiteout files in this nested VM, so
# pin the classic vfs graph driver. Writing the file is durable across boots.
sudo mkdir -p /etc/docker
echo '{ "storage-driver": "vfs" }' | sudo tee /etc/docker/daemon.json >/dev/null

log "Warm Bazel: build + test everything"
bazel build //...
bazel test //... --test_output=errors

log "Pre-pull database images into the snapshot"
# Start dockerd just long enough to cache images; start.sh owns the per-boot daemon.
if ! sudo docker info >/dev/null 2>&1; then
  sudo bash -c 'nohup dockerd >/var/log/dockerd.log 2>&1 &'
  for _ in $(seq 1 30); do sudo docker info >/dev/null 2>&1 && break; sleep 1; done
fi
if sudo docker info >/dev/null 2>&1; then
  sudo docker pull "$NEO4J_IMAGE"
  sudo docker pull "$QDRANT_IMAGE"
else
  echo "docker daemon not reachable during install; start.sh will pull on first boot" >&2
fi

log "install.sh complete"
