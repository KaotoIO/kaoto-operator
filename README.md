
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
 - Deploy Kaoto Custom Resource sample: `oc apply -f config/samples/_v1alpha1_kaoto.yaml`

 
### Via Operator Hub 
  - Install Kaoto operator catalog resource:  
    ```oc apply -f https://raw.githubusercontent.com/KaotoIO/kaoto-operator/main/catalogSource.yaml```
 - Install the Kaoto Operator from the Operator
 - Create Kaoto instance from the Kaoto Operator page
