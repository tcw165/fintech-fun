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

# ---------------------------------------------------------------------------
# Graphite CLI (`gt`)
#
# Team rule: all PR CRUD goes through Graphite. Cloud Agent VMs don't ship
# `gt`, so install it here. Node is provided by the base image via nvm; the
# GRAPHITE_AUTH_TOKEN secret is injected as an env var and picked up by `gt`
# automatically, so no `gt auth` step is needed at build time.
# ---------------------------------------------------------------------------
install_graphite_cli() {
	local nvm_prefix nvm_npm
	nvm_prefix=""
	# Resolve the newest nvm-managed node install (user-writable, on PATH via
	# ~/.bashrc). Fall back to whatever `npm` is on PATH.
	if [[ -d "$HOME/.nvm/versions/node" ]]; then
		nvm_prefix="$(ls -1d "$HOME"/.nvm/versions/node/*/ 2>/dev/null | sort -V | tail -1)"
		nvm_prefix="${nvm_prefix%/}"
	fi
	if [[ -n "$nvm_prefix" && -x "$nvm_prefix/bin/npm" ]]; then
		nvm_npm="$nvm_prefix/bin/npm"
	elif command -v npm >/dev/null 2>&1; then
		nvm_npm="$(command -v npm)"
		nvm_prefix="$(dirname "$(dirname "$nvm_npm")")"
	else
		log "npm not found; skipping gt install"
		return 0
	fi

	# Idempotent: skip the npm round-trip if gt is already installed at the
	# same version we would install.
	local target_version="1.8.6"
	if [[ -x "$nvm_prefix/bin/gt" ]] && "$nvm_prefix/bin/gt" --version 2>/dev/null | grep -qx "$target_version"; then
		log "gt $target_version already installed at $nvm_prefix/bin/gt"
	else
		log "installing @withgraphite/graphite-cli@$target_version into $nvm_prefix"
		"$nvm_npm" install -g --prefix "$nvm_prefix" --silent "@withgraphite/graphite-cli@$target_version"
	fi

	# Belt-and-suspenders: expose gt from /usr/local/bin so subprocesses with a
	# stripped PATH (no nvm) still find it. A symlink is not enough — the JS
	# entrypoint's `#!/usr/bin/env node` needs node on PATH too. A tiny
	# wrapper that hardcodes the node binary keeps gt working regardless.
	# Passwordless sudo is available on Cloud Agent VMs; if it isn't, we
	# fall back to the nvm-only install (still on PATH via ~/.bashrc).
	if command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
		local wrapper wrapper_path
		for wrapper in gt graphite; do
			wrapper_path="/usr/local/bin/$wrapper"
			# rm -f first so we replace any pre-existing symlink; without
			# this, `tee` would follow the symlink chain into node_modules
			# and clobber gt.js itself.
			sudo rm -f "$wrapper_path"
			sudo tee "$wrapper_path" >/dev/null <<-EOF
				#!/usr/bin/env bash
				exec "$nvm_prefix/bin/node" "$nvm_prefix/lib/node_modules/@withgraphite/graphite-cli/bin/gt.js" "\$@"
			EOF
			sudo chmod +x "$wrapper_path"
		done
	fi

	# Prove auth works. GRAPHITE_AUTH_TOKEN comes from the injected secret.
	if [[ -n "${GRAPHITE_AUTH_TOKEN:-}" ]]; then
		"$nvm_prefix/bin/gt" auth --token "$GRAPHITE_AUTH_TOKEN" >/dev/null 2>&1 || \
			log "gt auth returned non-zero (non-fatal; token env var still takes precedence)"
	else
		log "GRAPHITE_AUTH_TOKEN not set; gt is installed but not authenticated"
	fi
}

install_graphite_cli

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
