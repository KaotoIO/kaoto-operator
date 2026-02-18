# Kaoto Operator Architecture

## What is Kaoto?

[Kaoto](https://kaoto.io/) is a visual editor for Apache Camel integrations. It provides a web-based UI that allows users to create, edit, and manage Camel Routes, Kamelets, and Pipes through a drag-and-drop interface, backed by an integrated catalog of Camel components and Enterprise Integration Patterns (EIPs).

## What is the Kaoto Operator?

The Kaoto Operator is a Kubernetes operator that manages the full lifecycle of Kaoto instances inside a Kubernetes cluster. Rather than manually creating Deployments, Services, and Ingress resources, users declare a single `Kaoto` custom resource and the operator takes care of provisioning and maintaining all the underlying infrastructure.

The operator is built on top of the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) framework and follows the standard Kubernetes operator pattern: watch a Custom Resource, reconcile the desired state with the actual state, and report status back.

## High-Level Architecture

```
                          +--------------------------+
                          |    Kubernetes API Server  |
                          +--------+---------+-------+
                                   |         ^
                       Watch CRD   |         | Update Status
                       events      |         |
                          +--------v---------+-------+
                          |     Kaoto Operator        |
                          |  (kaoto-system namespace) |
                          |                           |
                          |  +---------+  +---------+ |
                          |  |  Leader  |  | Health  | |
                          |  | Election |  | Probes  | |
                          |  +---------+  +---------+ |
                          |                           |
                          |  +---------------------+  |
                          |  | KaotoReconciler     |  |
                          |  |                     |  |
                          |  |  1. ServiceAction   |  |
                          |  |  2. Ingress/Route   |  |
                          |  |  3. DeployAction    |  |
                          |  +---------------------+  |
                          +--------+------------------+
                                   |
                     Creates/Manages owned resources
                                   |
                    +--------------+--------------+
                    |              |              |
              +-----v----+  +-----v----+  +------v-----+
              | Service   |  | Ingress  |  | Deployment |
              | (port     |  | or Route |  | (Kaoto     |
              |  8080)    |  | (if      |  |  app       |
              |           |  |  needed) |  |  container)|
              +-----------+  +----------+  +------------+
```

## Custom Resource Definition

The operator introduces a single CRD:

| Field | Value |
|-------|-------|
| **API Group** | `designer.kaoto.io` |
| **Version** | `v1alpha1` |
| **Kind** | `Kaoto` |
| **Scope** | Namespaced |
| **Plural** | `kaotoes` |
| **Short name** | `kd` |
| **Categories** | `integration`, `camel` |

### Spec

```go
type KaotoSpec struct {
    Image   string       // Custom container image (default: quay.io/kaotoio/kaoto-app:stable)
    Ingress *IngressSpec // External access configuration (optional)
}

type IngressSpec struct {
    Host string // Hostname for Ingress/Route
    Path string // Path prefix for Ingress/Route
}
```

### Status

```go
type KaotoStatus struct {
    Phase              string             // "Ready" or "Error"
    Conditions         []metav1.Condition // Per-resource condition tracking
    ObservedGeneration int64              // Last reconciled generation
    Endpoint           string             // URL to access the Kaoto UI
}
```

**Status conditions** reported by the operator:

| Condition | Tracks |
|-----------|--------|
| `Reconcile` | Overall reconciliation outcome |
| `Deployment` | Kaoto pod deployment status |
| `Service` | Kubernetes Service creation |
| `Ingress` | Ingress or Route status |

### Example

```yaml
apiVersion: designer.kaoto.io/v1alpha1
kind: Kaoto
metadata:
  name: my-kaoto
spec:
  image: "quay.io/kaotoio/kaoto-app:stable"
  ingress:
    host: "kaoto.example.com"
    path: "/designer"
```

## Project Layout

```
kaoto-operator/
  api/designer/v1alpha1/         # CRD type definitions (KaotoSpec, KaotoStatus)
  cmd/main.go                    # CLI entrypoint (cobra)
  config/
    crd/                         # Generated CRD manifests
    manager/                     # Operator Deployment manifests
    rbac/                        # RBAC rules (ClusterRole, bindings)
    default/                     # Kustomize base overlay
    standalone/                  # Standalone installation overlay
  internal/controller/designer/  # Reconciliation logic and actions
  pkg/
    apply/                       # Server-side apply helpers
    client/                      # Composite Kubernetes client
    cmd/run/                     # "kaoto run" command implementation
    controller/                  # Manager and controller setup
    defaults/                    # Default values (image, intervals, finalizer)
    openshift/                   # OpenShift API detection
  test/                          # E2E tests and test support utilities
```

## Action-Based Reconciliation

The reconciliation loop is decomposed into discrete, composable **Actions**. Each action is responsible for a single resource type and implements the `Action` interface:

```go
type Action interface {
    Configure(context.Context, *client.Client, *builder.Builder) (*builder.Builder, error)
    Apply(context.Context, *ReconciliationRequest) error
    Cleanup(context.Context, *ReconciliationRequest) error
}
```

| Method | Purpose |
|--------|---------|
| `Configure` | Registers resource watches and predicates with the controller builder |
| `Apply` | Creates or updates the managed resource, sets status conditions |
| `Cleanup` | Deletes the managed resource when the Kaoto CR is being removed |

### Action Pipeline

Actions are executed sequentially in a fixed order during each reconciliation cycle:

```
1. ServiceAction      -->  Creates/updates a ClusterIP Service on port 8080
       |
2. IngressAction      -->  Creates/updates a Kubernetes Ingress    (vanilla K8s)
   or RouteAction      -->  Creates/updates an OpenShift Route      (OpenShift)
       |
3. DeployAction       -->  Creates/updates a Deployment running the Kaoto container
```

The choice between `IngressAction` and `RouteAction` is made once at startup based on cluster type detection.

On deletion, cleanup runs in **reverse order** to avoid dependency issues (the Deployment is removed before the Service).

## Reconciliation Flow

```
                  Kaoto CR event received
                          |
                          v
              Is DeletionTimestamp set?
                    /           \
                  No             Yes
                  |               |
          Add finalizer      Run Cleanup
          (if missing)       (reverse action order)
                  |               |
          Run all Actions    Remove finalizer
          sequentially            |
                  |            Return
                  v
          Any errors?
           /        \
         Yes         No
          |           |
   Phase="Error"   Phase="Ready"
          |           |
          +-----+-----+
                |
        Update Status
        (with conditions)
                |
           Return result
```

## Cluster Type Detection

At startup, the operator probes the Kubernetes API server for the `route.openshift.io/v1` API group:

- **Found** -> the cluster is OpenShift, the operator uses `RouteAction` for external access
- **Not found** -> the cluster is vanilla Kubernetes, the operator uses `IngressAction`

This detection happens once in `NewKaotoReconciler` and determines which action is added to the pipeline.

## Resources Created Per Kaoto Instance

For each `Kaoto` CR, the operator creates and manages these resources (all in the same namespace):

### 1. Service

- **Name**: same as the Kaoto CR
- **Port**: 8080 (TCP, named `http`)
- **Selector**: `app.kubernetes.io/name=kaoto`, `app.kubernetes.io/instance=<cr-name>`
- **Session affinity**: None
- **Publish not-ready addresses**: true

### 2. Ingress (vanilla Kubernetes) or Route (OpenShift)

Only created when `spec.ingress` is set on the Kaoto CR. Deleted automatically when the field is removed.

**Ingress** (vanilla K8s):
- NGINX ingress annotations for regex-based path rewriting
- Path defaults to `/<cr-name>(/|$)(.*)`
- `PathType`: ImplementationSpecific

**Route** (OpenShift):
- TLS termination at edge with HTTP-to-HTTPS redirect
- HAProxy rewrite annotation when a custom path is specified
- Endpoint reported as `https://` based on Route status

### 3. Deployment

- **Replicas**: 1
- **Image**: `spec.image` or `quay.io/kaotoio/kaoto-app:stable`
- **Image pull policy**: Always
- **Container port**: 8080
- **Probes**:
  - Readiness: HTTP GET `/` on port 8080 (5s initial delay, 10s period)
  - Liveness: HTTP GET `/` on port 8080 (5s initial delay, 10s period)
- **Resource requests**: 500m CPU, 600Mi memory
- **Environment**: `NAMESPACE` injected from pod metadata
- **Security context**:
  - Pod: `runAsNonRoot=true`, `seccompProfile=RuntimeDefault`
  - Container: `allowPrivilegeEscalation=false`, `runAsNonRoot=true`

### Resource Ownership

All created resources carry an `ownerReference` pointing back to the Kaoto CR. This provides:

- Automatic garbage collection if the Kaoto CR is deleted
- Controller-runtime watch filtering (the controller only reacts to resources it owns)

## Labeling Strategy

Every resource created by the operator carries these Kubernetes recommended labels:

```yaml
app.kubernetes.io/name: kaoto
app.kubernetes.io/instance: <kaoto-cr-name>
app.kubernetes.io/component: designer
app.kubernetes.io/part-of: kaoto
app.kubernetes.io/managed-by: kaoto-operator
```

Pod selection uses a minimal subset:

```yaml
app.kubernetes.io/name: kaoto
app.kubernetes.io/instance: <kaoto-cr-name>
```

## Server-Side Apply

The operator uses Kubernetes **server-side apply** for all resource mutations. This means:

- Resources are declared as typed apply-configurations rather than raw objects
- The operator acts as field manager `kaoto-operator`, so it only owns the fields it explicitly sets
- Other controllers or users can set non-conflicting fields without interference
- `Force: true` is used to resolve field ownership conflicts in favor of the operator

## Finalizer

The operator registers the finalizer `kaoto.io/finalizer` on every Kaoto CR. This ensures the operator gets a chance to clean up resources before the CR is removed from the API server, preventing orphaned Deployments, Services, or Ingress/Route resources.

## Operator Deployment

The operator itself runs as a Deployment in the `kaoto-system` namespace:

- **Image**: `quay.io/kaotoio/kaoto-operator:<version>`
- **Command**: `/kaoto run --leader-election`
- **Resource limits**: 500m CPU / 128Mi memory
- **Resource requests**: 10m CPU / 64Mi memory
- **Health probes**: `/healthz` (liveness) and `/readyz` (readiness) on port 8081
- **Metrics**: Prometheus metrics served on port 8080
- **Security**: non-root (UID 65532), read-only filesystem, dropped capabilities, no privilege escalation
- **Base image**: `gcr.io/distroless/static:nonroot`

### Leader Election

When running multiple replicas for high availability, the operator uses Kubernetes lease-based leader election with the identity `9aa9f118.kaoto.io`. Only the leader performs reconciliation; standby replicas take over automatically on leader failure.

## RBAC

The operator's ClusterRole grants these permissions:

| API Group | Resources | Verbs |
|-----------|-----------|-------|
| `""` (core) | services | get, list, watch, create, update, patch, delete |
| `apps` | deployments | get, list, watch, create, update, patch, delete |
| `designer.kaoto.io` | kaotoes | get, list, watch, create, update, patch, delete |
| `designer.kaoto.io` | kaotoes/status | get, update, patch |
| `designer.kaoto.io` | kaotoes/finalizers | update |
| `networking.k8s.io` | ingresses | get, list, watch, create, update, patch, delete |
| `route.openshift.io` | routes | get, list, watch, create, update, patch, delete |

## How the Operator Links With Kaoto

The operator acts as the **deployment and lifecycle manager** for Kaoto on Kubernetes. The relationship is:

1. **Kaoto** is a standalone web application (container image `quay.io/kaotoio/kaoto-app`) that provides a visual editor for Apache Camel integrations. It runs as a single HTTP server on port 8080.

2. **The operator** takes that container image and wraps it in the Kubernetes primitives needed to run it reliably: a Deployment for pod management, a Service for network access, and an Ingress or Route for external exposure.

3. **The CRD** (`Kaoto`) provides a simplified, domain-specific interface. Instead of writing and maintaining three or more Kubernetes manifests, users create a single resource with two optional fields (`image` and `ingress`), and the operator handles the rest.

4. **Lifecycle management**: the operator continuously reconciles the desired state. If someone deletes the underlying Deployment or modifies the Service, the operator detects the drift and restores the correct state. When the Kaoto CR is deleted, the operator cleans up all owned resources.

5. **Status reporting**: the operator surfaces the Kaoto application's access URL in `.status.endpoint`, so users can discover where their Kaoto instance is available without inspecting multiple resources.

```
User creates Kaoto CR
        |
        v
Operator reconciles
        |
        +---> Service (internal networking)
        +---> Ingress/Route (external access)
        +---> Deployment (runs kaoto-app container)
        |
        v
Kaoto UI accessible at .status.endpoint
```

## Release Process

The Kaoto ecosystem follows a coordinated release sequence: the main Kaoto application is released first, and the operator release follows once the new application image is available.

### Kaoto Application Release (upstream)

The [Kaoto repository](https://github.com/KaotoIO/kaoto) release pipeline is triggered manually via `workflow_dispatch` with two inputs:

| Input | Type | Description |
|-------|------|-------------|
| `tag_version` | string | The version to release (e.g. `2.4.0`) |
| `stable` | boolean | Whether to tag the container image as `stable` (default: `true`) |

The pipeline runs three sequential jobs:

```
tag-and-release
      |
      v
 npm-release          Publishes @kaoto/kaoto to NPM
      |
      v
container-image-release
      |
      +---> pushes quay.io/kaotoio/kaoto-app:<version>
      +---> pushes quay.io/kaotoio/kaoto-app:stable   (if stable=true)
```

1. **tag-and-release** -- Creates a Git tag and a GitHub release for the version.
2. **npm-release** -- Builds the `@kaoto/kaoto` library package and publishes it to the NPM registry using OIDC authentication.
3. **container-image-release** -- Builds the Kaoto web application container image, pushes it to `quay.io/kaotoio/kaoto-app` tagged with the version number, and optionally tags it as `stable`.

The `stable` tag is the key link to the operator: the operator defaults to `quay.io/kaotoio/kaoto-app:stable`, so tagging a Kaoto release as stable makes it the version that all new operator-managed Kaoto instances will run.

### Operator Release

Once the new Kaoto application image is available, the [operator release pipeline](https://github.com/KaotoIO/kaoto-operator/blob/main/.github/workflows/release-pipeline.yml) is triggered manually with these inputs:

| Input | Type | Description |
|-------|------|-------------|
| `default_bump` | choice | Semver bump type: `major`, `minor`, or `patch` |
| `operator_version` | string | The operator version to release |
| `replace_operator_version` | string | The previous operator version this release replaces (for OLM upgrade chains) |
| `supported_openshift_range` | string | Minimum supported OpenShift version (default: `v4.13`) |
| `latest` | boolean | Whether to also tag the operator image as `latest` |

The pipeline runs four jobs:

```
build-and-push
      |
      v
tag-and-release
      |
      +---------------------------+
      |                           |
      v                           v
prepare-k8s-operator-branch   prepare-openshift-operator-branch
```

#### 1. build-and-push

- Checks out the source, sets up Go, builds the operator binary, and runs tests
- Builds and pushes the operator container image to `quay.io/kaotoio/kaoto-operator:<version>`
- Optionally tags the image as `latest`
- Generates the OLM (Operator Lifecycle Manager) bundle using `make bundle`
- Enriches the bundle CSV with the `replaces` field pointing to the previous version
- Uploads the bundle as a build artifact

#### 2. tag-and-release

- Creates a Git tag for the version
- Creates a GitHub release with the OLM bundle attached as a release artifact

#### 3. prepare-k8s-operator-branch

- Checks out the [community-operators](https://github.com/k8s-operatorhub/community-operators) repository
- Copies the generated bundle into the appropriate version directory
- Opens a pull request to publish the operator on **OperatorHub.io** (Kubernetes)

#### 4. prepare-openshift-operator-branch

- Checks out the [community-operators-prod](https://github.com/redhat-openshift-ecosystem/community-operators-prod) repository
- Copies the generated bundle and adds OpenShift version compatibility annotations
- Opens a pull request to publish the operator on the **Red Hat OpenShift ecosystem catalog**

### Pre-Release: Syncing the Community Operator Forks

Before triggering the operator release, it is **critical** to synchronize the KaotoIO forks of the community operator repositories with their upstream origins:

- [KaotoIO/community-operators](https://github.com/KaotoIO/community-operators) (fork of [k8s-operatorhub/community-operators](https://github.com/k8s-operatorhub/community-operators))
- [KaotoIO/community-operators-prod](https://github.com/KaotoIO/community-operators-prod) (fork of [redhat-openshift-ecosystem/community-operators-prod](https://github.com/redhat-openshift-ecosystem/community-operators-prod))

The release workflow's `prepare-k8s-operator-branch` and `prepare-openshift-operator-branch` jobs check out these forks, create a new branch with the bundle, and open a pull request against the upstream repository. If the forks are out of date with their upstream, the PR branch will be based on stale history, which can cause merge conflicts or CI failures on the upstream PR.

**Sync procedure** (for each fork):

```bash
# Example for community-operators
git clone git@github.com:KaotoIO/community-operators.git
cd community-operators
git remote add upstream https://github.com/k8s-operatorhub/community-operators.git
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
```

Or use the GitHub UI: on each fork's page, click **"Sync fork"** to bring it up to date with the upstream default branch.

This must be done for both forks before launching the operator release pipeline.

### Release Coordination

The two release pipelines are connected through the container image tag:

```
Kaoto repo                              Operator repo
----------                              -------------

Release v2.4.0                          Release operator v0.0.5
    |                                        |
    v                                        v
Push kaoto-app:2.4.0                    Push kaoto-operator:0.0.5
Push kaoto-app:stable  <--- used by --- Operator default image:
    |                                     quay.io/kaotoio/kaoto-app:stable
    v                                        |
Image available on quay.io                   v
                                        OLM bundle submitted to
                                        OperatorHub + OpenShift catalog
```

The sequence matters:

1. The Kaoto application must be released first so the `stable` container image exists on the registry.
2. The operator release follows. Since the operator deploys `kaoto-app:stable` by default, any new Kaoto instances created after the application release will automatically pick up the latest version.
3. The operator release also publishes OLM bundles to OperatorHub (Kubernetes) and the Red Hat OpenShift ecosystem, making the operator installable through the standard catalog UIs on both platforms.

Users who need to pin a specific Kaoto version can override the image in the CR spec:

```yaml
spec:
  image: "quay.io/kaotoio/kaoto-app:2.4.0"
```

## Installation

### From the repository

```bash
kubectl apply -k https://github.com/KaotoIO/kaoto-operator/config/standalone
```

### Creating a Kaoto instance

```yaml
apiVersion: designer.kaoto.io/v1alpha1
kind: Kaoto
metadata:
  name: kaoto
spec:
  ingress:
    host: kaoto.example.com
```

Check the status:

```bash
kubectl get kaoto
# NAME    PHASE   ENDPOINT
# kaoto   Ready   https://kaoto.example.com/
```
