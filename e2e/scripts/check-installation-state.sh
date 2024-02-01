#!/usr/bin/env bash
set -euo pipefail

main() {
    echo "installations"
    kubectl get installations
    kubectl describe installations
    echo "charts"
    kubectl get charts -A
    kubectl describe charts -A
    echo "pods"
    kubectl get pods -A
    kubectl rollout restart deployment/kotsadm -n kotsadm
    sleep 30
}

export EMBEDDED_CLUSTER_METRICS_BASEURL="https://staging.replicated.app"
export KUBECONFIG=/root/.config/embedded-cluster/etc/kubeconfig
export PATH=$PATH:/root/.config/embedded-cluster/bin
main
