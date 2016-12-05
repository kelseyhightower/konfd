# Deployment Guide

## Kubernetes

konfd should be deployed using a replicaset with a replica count set to 1 to avoid proccessing the same templates multiple times. By default konfd will process all templates for all namespaces, but it's also possible to limit template processing to specific namespaces by specifying the `-namespace` flag multiple times. Example:

```
containers:
  - name: konfd
    image: "gcr.io/hightowerlabs/konfd:v0.0.2"
    args:
      - "-namespace=default"
      - "-namespace=dev"
```  

### Create the konfd replicaset

```
kubectl create -f replicasets/konfd.yaml
```
