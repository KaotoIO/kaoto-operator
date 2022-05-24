
This repository consist of Kaoto operator and other kubernetes resources that helps to run Kaoto in the Kubernetes cluster.

# Kaoto operator
The Kubernetes operator that manages Kaoto instance within the kubernetes clusters. 


# Kubernetes resources
Multiresource yaml files to deploy to plain kubernetes. 




## Install Kaoto

### Plain Kubernetes (Minikube)

- Install and run a Minikube instance with `ingress` addon enabled. 
- Install Kaoto from the multi-resource yaml 
  ```kubect apply -f https://raw.githubusercontent.com/KaotoIO/kaoto-operator/main/kubernetes/kaoto.yaml``` 
  -  this will create kaoto namespace, install kaoto and create ingress with `kaoto.local` address 
- add record with actuall ip of the cluster to `/etc/hosts` :
  
  ``` (minikube ip) kaoto.local kaoto.backend.local```
- Kaoto should be accessible at `http://kaoto.local` 

