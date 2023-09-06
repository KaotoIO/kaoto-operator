
This repository consist of Kaoto operator and other Kubernetes resources that helps to run Kaoto in the Kubernetes cluster.

# Kaoto operator
The Kubernetes operator that manages Kaoto instance within the Kubernetes clusters. 


# Kubernetes resources
Multiresource yaml files to deploy to plain kubernetes. 



## Install Kaoto

### Plain Kubernetes (Minikube)
- Install and run a Minikube instance with `ingress` addon enabled. 
- Install Kaoto from the multi-resource yaml 
  ```kubectl apply -k https://github.com/KaotoIO/kaoto-operator//config/standalone``` 
  - this will create `kaoto-system` namespace and install Kaoto Operatorand 
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
1. Start minikube win ingress controller enabled: `minikube start --addons ingress`
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
