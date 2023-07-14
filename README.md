
This repository consist of Kaoto operator and other Kubernetes resources that helps to run Kaoto in the Kubernetes cluster.

# Kaoto operator
The Kubernetes operator that manages Kaoto instance within the Kubernetes clusters. 


# Kubernetes resources
Multiresource yaml files to deploy to plain kubernetes. 



## Install Kaoto

### Plain Kubernetes (Minikube)

- Install and run a Minikube instance with `ingress` addon enabled. 
- Install Kaoto from the multi-resource yaml 
  ```kubect apply -f https://raw.githubusercontent.com/KaotoIO/kaoto-operator/main/kubernetes/kaoto.yaml``` 
  -  this will create `kaoto` namespace, install Kaoto and create Ingress with `kaoto.local` address 
- Add record with actual ip of the cluster to `/etc/hosts`:  
  ```(minikube ip) kaoto.local kaoto.backend.local```
- Kaoto should be accessible at `http://kaoto.local` 

### Using the Operator
 - Clone `kaoto-operator` repository 
 - Run `make deploy` which creates `kaoto-operator` project and deploy all necessary resources as Kaoto CRD and necessary `serviceaccounts`
 - Deploy Kaoto Custom Resource sample: `kubectl apply -f config/samples/designer_v1alpha1_kaoto.yaml`

 
### Via Operator Hub 
  - Install Kaoto operator catalog resource:  
    ```oc apply -f https://raw.githubusercontent.com/KaotoIO/kaoto-operator/main/catalogSource.yaml```
 - Install the Kaoto Operator from the Operator
 - Create Kaoto instance from the Kaoto Operator page

## Local development

### Run Operator inside the cluster
1. Start minikube win ingress controller enabled: `minikube start --addons ingress`
2. Point docker to minikube internal registry: `eval $(minikube -p minikube docker-env)`
3. Build the Operator: `make build`
4. Build the Operator Image: `make docker-build`
5. Deploy Operator: `make deploy`
6. Create sample Kaoto CR: `kubectl apply -f config/samples/designer_v1alpha1_kaoto.yaml`
7. (Optional) Undeploy everything: `make undeploy`

### Run locally outside the cluster
1. Start minikube win ingress controller enabled: `minikube start --addons ingress`
2. Run operator locally: `make install run`
3. Create sample Kaoto CR: `kubectl apply -f config/samples/designer_v1alpha1_kaoto.yaml`
4. (Optional) Undeploy Kaoto: `kubectl delete kaoto kaoto-sample`
