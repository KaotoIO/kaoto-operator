#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"
OUTPUT_DIR="${PROJECT_ROOT}/pkg/client/kaoto"

mkdir -p "${OUTPUT_DIR}"

"${PROJECT_ROOT}"/bin/applyconfiguration-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${OUTPUT_DIR}/applyconfiguration" \
  --output-pkg=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/applyconfiguration \
  github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1

"${PROJECT_ROOT}"/bin/client-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${OUTPUT_DIR}/clientset" \
  --output-pkg=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/clientset \
  --plural-exceptions="Kaoto:Kaotoes" \
  --input=designer/v1alpha1 \
  --clientset-name "versioned" \
  --input-base=github.com/kaotoIO/kaoto-operator/api \
  --apply-configuration-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/applyconfiguration

"${PROJECT_ROOT}"/bin/lister-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${OUTPUT_DIR}/listers" \
  --output-pkg=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/listers \
  --plural-exceptions="Kaoto:Kaotoes" \
  github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1

"${PROJECT_ROOT}"/bin/informer-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${OUTPUT_DIR}/informers" \
  --output-pkg=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/informers \
  --plural-exceptions="Kaoto:Kaotoes" \
  --versioned-clientset-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/clientset/versioned \
  --listers-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/listers \
  github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1
