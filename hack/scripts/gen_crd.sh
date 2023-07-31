#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"

"${PROJECT_ROOT}"/bin/controller-gen \
  rbac:roleName=kaoto-operator \
  crd \
  paths="./..." output:crd:artifacts:config="${PROJECT_ROOT}/config/crd/bases"
