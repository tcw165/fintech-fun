#!/usr/bin/env bash
# Cloud Agent install hook.
#
# Cursor's environment.json points its install step at .cursor/install.sh, so
# this file must exist and exit 0 for the environment build to succeed.
#
# The Cloud Agent base image already ships bazelisk/bazel, go, docker, and gh.
# All we need to do here is warm project-level caches so the first `bazel test`
# in an agent session is quick, and surface obvious misconfiguration early.
#
# Keep this script self-contained: no `just`, `minikube`, or `kubectl` — those
# are host-only concerns (see README.md) and are not present in the base image.

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

log() { printf '[.cursor/install.sh] %s\n' "$*"; }

log "workspace: $repo_root"

if ! command -v bazelisk >/dev/null 2>&1 && ! command -v bazel >/dev/null 2>&1; then
	log "bazel/bazelisk not found on PATH; skipping cache warmup"
	exit 0
fi

bazel_bin="$(command -v bazelisk || command -v bazel)"
log "using $bazel_bin ($($bazel_bin --version 2>/dev/null | head -1 || echo unknown))"

# Resolve MODULE.bazel deps and materialize external repos referenced by
# BUILD graph. `bazel fetch //...` is enough to prime rules_go, gazelle, the
# Go SDK download, and the go_deps set without actually compiling anything.
log "priming bazel external repositories (bazel fetch //...)"
"$bazel_bin" fetch //... >/dev/null

# Pre-download Go module dependencies too, so non-bazel tooling (IDE indexers,
# `go build ./...`, etc.) doesn't stall on the first invocation.
if command -v go >/dev/null 2>&1; then
	log "priming go module cache (go mod download)"
	GOFLAGS="-mod=mod" go mod download all || log "go mod download reported warnings (non-fatal)"
fi

log "done"
