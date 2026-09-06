# Local minikube cluster. Profile, kube context, and namespace are all `fintech-fun`.
# Homebrew first: a leftover /usr/local/bin/minikube on Mac is often a Linux binary
# (`exec format error`), which breaks `just down` / `just up` / `just healthz`.
export PATH := "/opt/homebrew/bin:/usr/local/bin:" + env_var("PATH")

profile := "fintech-fun"
ns := "fintech-fun"
image_dir := "/tmp/fintech-fun-images"

default:
    @just --list

# Host tools for the Docker-driver minikube. On Mac this means Docker Desktop is running.
deps:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v docker >/dev/null; then
      echo "docker is required. On Mac: install Docker Desktop and ensure docker is on PATH." >&2
      exit 1
    fi
    if ! docker info >/dev/null 2>&1; then
      echo "docker daemon is not running. On Mac, start Docker Desktop, then retry." >&2
      exit 1
    fi

# Start minikube (Docker driver). Safe to re-run if the profile already exists.
cluster-start: deps
    minikube start --profile={{profile}} --driver=docker

# Bazel-build Linux binaries, docker-build images, load them into the minikube profile.
cluster-images: deps
    #!/usr/bin/env bash
    set -euo pipefail
    profile="{{profile}}"
    image_dir="{{image_dir}}"
    minikube_bin="$(command -v minikube)"
    arch=""
    if "$minikube_bin" status -p "$profile" >/dev/null 2>&1; then
      arch="$("$minikube_bin" -p "$profile" ssh -- uname -m | tr -d '\r\n')"
    fi
    if [[ -z "$arch" ]]; then
      case "$(uname -m)" in
        arm64|aarch64) arch=aarch64 ;;
        x86_64|amd64) arch=x86_64 ;;
      esac
    fi
    case "$arch" in
      aarch64|arm64) platform="@rules_go//go/toolchain:linux_arm64" ;;
      x86_64|amd64) platform="@rules_go//go/toolchain:linux_amd64" ;;
      *) echo "unsupported minikube arch: $arch" >&2; exit 1 ;;
    esac
    bazel_flags=(--platforms="$platform" --@rules_go//go/config:pure)
    bazel build "${bazel_flags[@]}" //api_server/cmd:api_server //ingest_jobs/corporate_actions:corporate_actions
    rm -rf "$image_dir"
    mkdir -p "$image_dir/api-server" "$image_dir/corporate-actions"
    cp "$(bazel cquery --noshow_progress --ui_event_filters=-info "${bazel_flags[@]}" --output=files //api_server/cmd:api_server)" "$image_dir/api-server/api_server"
    cp "$(bazel cquery --noshow_progress --ui_event_filters=-info "${bazel_flags[@]}" --output=files //ingest_jobs/corporate_actions:corporate_actions)" "$image_dir/corporate-actions/corporate_actions"
    file "$image_dir/api-server/api_server"
    # Build inside minikube's Docker daemon so Kubernetes picks up the new tag
    # (minikube image load does not replace an in-use :local tag).
    eval "$("$minikube_bin" docker-env -p "$profile")"
    docker build --no-cache -f api_server/Dockerfile -t fintech-fun/api-server:local "$image_dir/api-server"
    docker build --no-cache -f ingest_jobs/corporate_actions/Dockerfile -t fintech-fun/corporate-actions:local "$image_dir/corporate-actions"

# Apply the local overlay (Neo4j, Qdrant, api-server, suspended CronJob).
cluster-apply:
    kubectl --context {{profile}} apply -k infra/k8s/overlays/local

# Start cluster, load images, apply manifests.
up: cluster-start cluster-images cluster-apply

# Stop the VM. Data in emptyDir is gone next start. Needs Docker Desktop on Mac.
down: deps
    #!/usr/bin/env bash
    set -euo pipefail
    if ! minikube version >/dev/null 2>&1; then
      echo "minikube is not runnable (wrong arch or missing). Prefer Homebrew: /opt/homebrew/bin/minikube" >&2
      exit 1
    fi
    minikube stop -p "{{profile}}"

# Delete the profile entirely.
cluster-delete: deps
    minikube delete -p {{profile}}

# Hit api-server /healthz through minikube.
healthz:
    curl -sS "$(minikube service -p {{profile}} -n {{ns}} api-server --url)/healthz"
    @echo

# Account-screen fold through minikube. Example: just fold square 10
fold q qty:
    curl -sS "$(minikube service -p {{profile}} -n {{ns}} api-server --url)/fold?q={{q}}&qty={{qty}}"
    @echo

# Open k9s on this cluster in the fintech-fun namespace.
k9s:
    k9s --context {{profile}} --namespace {{ns}}

# Re-fetch the tracker and MERGE new Event ids. Safe to re-run after new trading days.
# waiting rows stay; a later cashed_out/now_different_stock is a new Event.
refresh:
    bazel run //ingest_jobs/corporate_actions -- refresh
