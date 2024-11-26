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
