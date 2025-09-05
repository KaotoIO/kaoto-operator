<h1 align="center">
  <img src="https://github.com/KaotoIO/kaoto.io/blob/main/assets/media/logo-kaoto.png" alt="Kaoto">
</h1>

<p align=center>
  <a href="https://github.com/KaotoIO/kaoto-operator/blob/main/LICENSE"><img src="https://img.shields.io/github/license/KaotoIO/kaoto-operator?color=blue&style=for-the-badge" alt="License"/></a>
  <a href="https://www.youtube.com/@KaotoIO"><img src="https://img.shields.io/badge/Youtube-Follow-brightgreen?color=red&style=for-the-badge" alt="Youtube"" alt="Follow on Youtube"></a>
  <a href="https://camel.zulipchat.com/#narrow/stream/441302-kaoto"><img src="https://img.shields.io/badge/zulip-join_chat-brightgreen?color=yellow&style=for-the-badge" alt="Zulip"/></a>
  <a href="https://kaoto.io"><img src="https://img.shields.io/badge/Kaoto.io-Visit-white?color=indigo&style=for-the-badge" alt="Zulip"/></a>
</p><br/>

<h2 align="center">Kaoto - The Integration Designer for <a href="https://camel.apache.org">Apache Camel</a></h2>

<p align="center">
  <a href="https://kaoto.io/docs/installation">Documentation</a> | 
  <a href="https://kaoto.io/workshop/">Workshops</a> | 
  <a href="https://kaoto.io/contribute/">Contribute</a> | 
  <a href="https://camel.zulipchat.com/#narrow/stream/441302-kaoto">Chat</a>
</p>

# Kaoto
Kaoto is a visual editor for Apache Camel integrations. It offers support in creating and editing Camel Routes, Kamelets and Pipes. Kaoto also has a built-in catalog with available Camel components, Enterprise Integration Patterns and Kamelets provided by the Apache Camel community.

Have a quick look at our online demo instance:
https://kaotoio.github.io/kaoto/

# Kaoto operator
The Kubernetes operator that manages Kaoto instance within the Kubernetes clusters. 

# Kubernetes resources
Multiresource yaml files to deploy to plain kubernetes. 

## Install Kaoto

### Plain Kubernetes (Minikube)
- Install and run a Minikube instance with `ingress` addon enabled. 
- Install Kaoto from the multi-resource yaml 
  ```kubectl apply -k https://github.com/KaotoIO/kaoto-operator/config/standalone``` 
  - this will create `kaoto-system` namespace and install Kaoto Operator 
- Create sample Kaoto CR
  ```kubectl apply -f https://raw.githubusercontent.com/KaotoIO/kaoto-operator/main/config/samples/designer.yaml```
- Waith the the ingrees admits the endoint
  ```  ➜ k get kaotos.designer.kaoto.io -w
  NAME       PHASE   ENDPOINT
  designer   Ready   http://192.168.49.2/designer/
  ```
- Kaoto should be accessible at `http:/$(minikube ip)/designer`

### Using the Operator
 - Clone `kaoto-operator` repository 
 - Run `make deploy` which creates `kaoto-system` project and deploy all necessary resources
 - Deploy Kaoto Custom Resource sample: `kubectl apply -f config/samples/designer.yaml`

## Local development

### Run Operator inside the cluster
1. Start minikube with ingress controller enabled: `minikube start --addons ingress`
2. Point docker to minikube internal registry: `eval $(minikube -p minikube docker-env)`
3. Build the Operator: `make build`
4. Build the Operator Image: `make docker-build`
5. Deploy Operator: `make deploy`
6. Create sample Kaoto CR: `kubectl apply -f config/samples/designer.yaml`
7. (Optional) Undeploy everything: `make undeploy`

### Run locally outside the cluster
1. Start minikube win ingress controller enabled: `minikube start --addons ingress`
2. Run operator locally: `make run/local`
3. Create sample Kaoto CR: `kubectl apply -f config/samples/designer.yaml`
4. (Optional) Undeploy Kaoto: `kubectl delete kaoto kaoto-demo`

## Build and Push to Custom Registry

To build and push the operator and bundle images to your own registry:

### Prerequisites
- Docker/Podman logged into your registry
- Operator SDK will be automatically installed via `make bundle`

### Build Process
1. **Set environment variables:**
   ```bash
   export MY_REGISTRY="your-registry.com"
   export MY_REPO="your-username/kaoto"
   export IMAGE_TAG_BASE="${MY_REGISTRY}/${MY_REPO}-operator"
   export VERSION="0.0.5"
   ```

2. **Generate bundle:**
   ```bash
   IMG=${IMAGE_TAG_BASE}:latest VERSION=${VERSION} make bundle
   ```

3. **Build and push operator and bundle images:**
   ```bash
   IMG=${IMAGE_TAG_BASE}:latest BUNDLE_IMG=${IMAGE_TAG_BASE}-bundle:v${VERSION} make docker-build docker-push bundle-build bundle-push
   ```

### Deploy Options

#### Option 1: Direct Deployment
```bash
IMG=${IMAGE_TAG_BASE}:latest make deploy
kubectl apply -f config/samples/designer.yaml
```

#### Option 2: Deploy via OLM Bundle
1. **Install OLM (if not present):**
   ```bash
   ./bin/operator-sdk olm install
   ```

2. **Deploy bundle using operator-sdk:**
   ```bash
   ./bin/operator-sdk run bundle ${IMAGE_TAG_BASE}-bundle:v${VERSION}
   ```

3. **Create Kaoto instance:**
   ```bash
   kubectl apply -f config/samples/designer.yaml
   ```

4. **Cleanup bundle (when done):**
   ```bash
   ./bin/operator-sdk cleanup kaoto-operator
   ```
