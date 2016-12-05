# konfd

Manage application configuration using Kubernetes secrets, configmaps, and Go templates.

## Usage

Create configmaps and secrets:

```
kubectl create secret generic vault-secrets \
  --from-literal 'mysql_password=v@ulTi$d0p3'
```

```
kubectl create configmap vault-configs \
  --from-literal 'default_lease_ttl=768h' \
  --from-literal 'mysql_username=vault'
```

Create the `template` configmap:

```
kubectl create -f configmaps/template.yaml
```

Process all konfd templates in all namespaces:

```
konfd
```

```
2016/12/04 23:02:51 Starting konfd...
2016/12/04 23:02:52 Syncing templates complete. Next sync in 60 seconds.
```

Review the results:

```
kubectl get configmaps vault -o yaml
```

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: vault
  namespace: default
data:
  server.hcl: |
    default_lease_ttl = 768h
    backend "mysql" {
      username = "vault"
      password = "v@ulTi$d0p3"
      tls_ca_file = "/etc/tls/mysql-ca.pem"
    }
```

### Testing with noop mode

```
konfd -onetime -noop -namespace default -configmap template
```
