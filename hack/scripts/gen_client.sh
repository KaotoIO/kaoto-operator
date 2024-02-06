#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"
TMP_DIR=$( mktemp -d -t kaoto-client-gen-XXXXXXXX )

mkdir -p "${TMP_DIR}/client"
mkdir -p "${PROJECT_ROOT}/pkg/client/kaoto"

"${PROJECT_ROOT}"/bin/applyconfiguration-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1 \
  --output-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/applyconfiguration

"${PROJECT_ROOT}"/bin/client-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --plural-exceptions="Kaoto:Kaotoes" \
  --input=designer/v1alpha1 \
  --clientset-name "versioned" \
  --input-base=github.com/kaotoIO/kaoto-operator/api \
  --apply-configuration-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/applyconfiguration \
  --output-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/clientset

"${PROJECT_ROOT}"/bin/lister-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --plural-exceptions="Kaoto:Kaotoes" \
  --input-dirs=github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1 \
  --output-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/listers

"${PROJECT_ROOT}"/bin/informer-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --plural-exceptions="Kaoto:Kaotoes" \
  --input-dirs=github.com/kaotoIO/kaoto-operator/api/designer/v1alpha1 \
  --versioned-clientset-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/clientset/versioned \
  --listers-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/listers \
  --output-package=github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/informers

cp -R "${TMP_DIR}"/client/github.com/kaotoIO/kaoto-operator/pkg/client/kaoto/* "${PROJECT_ROOT}"/pkg/client/kaoto