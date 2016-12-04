# konfd

Manage application configuration using Kubernetes secrets, configmaps, and Go templates.

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

```
kubectl create configmap vault-template --from-file vault.hcl --dry-run -o yaml \
  > vault-template.yaml
```
