#!/usr/bin/env bash


set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
SCRIPT_ROOT="${SCRIPT_DIR}/.."
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
bash "${CODEGEN_PKG}"/generate-groups.sh \
  "deepcopy,client,informer,lister" \
  github.com/ngsin/demo-controller/pkg/client \
  github.com/ngsin/demo-controller/pkg/apis \
  demoapi:v1alpha1 \
  --output-base "${SCRIPT_ROOT}" \
  --go-header-file "${SCRIPT_DIR}"/boilerplate.go.txt