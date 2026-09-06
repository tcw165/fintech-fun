# Local minikube cluster. Profile, kube context, and namespace are all `fintech-fun`.
profile := "fintech-fun"
ns := "fintech-fun"
image_dir := "/tmp/fintech-fun-images"

default:
    @just --list

# Start minikube (Docker driver). Safe to re-run if the profile already exists.
cluster-start:
    minikube start --profile={{profile}} --driver=docker

# Bazel-build binaries, docker-build images, load them into the minikube profile.
cluster-images:
    bazel build //api_server/cmd:api_server //ingest_jobs/corporate_actions:corporate_actions
    mkdir -p {{image_dir}}/api-server {{image_dir}}/corporate-actions
    cp "$(bazel cquery --noshow_progress --ui_event_filters=-info --output=files //api_server/cmd:api_server)" {{image_dir}}/api-server/api_server
    cp "$(bazel cquery --noshow_progress --ui_event_filters=-info --output=files //ingest_jobs/corporate_actions:corporate_actions)" {{image_dir}}/corporate-actions/corporate_actions
    docker build -f api_server/Dockerfile -t fintech-fun/api-server:local {{image_dir}}/api-server
    docker build -f ingest_jobs/corporate_actions/Dockerfile -t fintech-fun/corporate-actions:local {{image_dir}}/corporate-actions
    minikube image load -p {{profile}} fintech-fun/api-server:local
    minikube image load -p {{profile}} fintech-fun/corporate-actions:local

# Apply the local overlay (Neo4j, Qdrant, api-server, suspended CronJob).
cluster-apply:
    kubectl --context {{profile}} apply -k infra/k8s/overlays/local

# Start cluster, load images, apply manifests.
up: cluster-start cluster-images cluster-apply

# Stop the VM. Data in emptyDir is gone next start.
down:
    minikube stop -p {{profile}}

# Delete the profile entirely.
cluster-delete:
    minikube delete -p {{profile}}

# Hit api-server /healthz through minikube.
healthz:
    curl -sS "$(minikube service -p {{profile}} -n {{ns}} api-server --url)/healthz"
    @echo

# Open k9s on this cluster.
k9s:
    k9s --context {{profile}}
