#!/bin/sh

if [ $# -ne 3 ]; then
    echo "project root, codegen version and is expected"
fi

PROJECT_ROOT="$1"
CODEGEN_VERSION="$2"
CONTROLLER_TOOLS_VERSION="$3"

GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/applyconfiguration-gen@"${CODEGEN_VERSION}"
GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/client-gen@"${CODEGEN_VERSION}"
GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/lister-gen@"${CODEGEN_VERSION}"
GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/informer-gen@"${CODEGEN_VERSION}"
GOBIN="${PROJECT_ROOT}/bin" go install sigs.k8s.io/controller-tools/cmd/controller-gen@"${CONTROLLER_TOOLS_VERSION}"