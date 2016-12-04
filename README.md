# konfd

Manage application configuration using Kubernetes secrets, configmaps, and Go templates.

## Usage

Create configmaps and secrets:

```
kubectl create secret generic vault \
  --from-literal 'mysql.password=v@ulTi$d0p3'
```

```
kubectl create configmap vault \
  --from-literal 'default_lease_ttl=768h' \
  --from-literal 'max_lease_ttl=768h' \
  --from-literal 'mysql.username=vault' \
  --from-literal 'mysql.address=23.12.4.3:3306' \
  --from-literal 'mysql.database=vault' \
  --from-literal 'mysql.table=vault'
```

Create the `vault-template` configmap:

```
kubectl create -f configmaps/vault-template.yaml
```

### Testing with noop mode

```
konfd --noop --namespace default --configmap-name vault-template
```

```
{
  "apiVersion": "v1",
  "data": {
    "mysql.password": "dkB1bFRpJGQwcDM=",
    "server.hcl": "ZGVmYXVsdF9sZWFzZV90dGwgPSA3NjhoCm1heF9sZWFzZV90dGwgPSA3NjhoCgpsaXN0ZW5lciAidGNwIiB7CiAgYWRkcmVzcyA9ICIwLjAuMC4wOjgyMDAiCiAgdGxzX2NlcnRfZmlsZSA9ICIvZXRjL3Rscy9zZXJ2ZXIucGVtIgogIHRsc19rZXlfZmlsZSA9ICIvZXRjL3Rscy9zZXJ2ZXIua2V5Igp9CgpiYWNrZW5kICJteXNxbCIgewogIHVzZXJuYW1lID0gInZhdWx0IgogIHBhc3N3b3JkID0gInZAdWxUaSRkMHAzIgogIGFkZHJlc3MgPSAiMjMuMTIuNC4zOjMzMDYiCiAgZGF0YWJhc2UgPSAidmF1bHQiCiAgdGFibGUgPSAidmF1bHQiCiAgdGxzX2NhX2ZpbGUgPSAiL2V0Yy90bHMvbXlzcWwtY2EucGVtIgp9Cg=="
  },
  "kind": "Secret",
  "metadata": {
    "name": "vault",
    "namespace": "default"
  },
  "type": "Opaque"
}
```

Notice the `server.hcl` has been added to the existing `vault` secret.
