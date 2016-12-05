# konfd

Manage application configuration using Kubernetes secrets, configmaps, and Go templates.

## Usage

Create configmaps and secrets:

```
kubectl create secret generic vault-secrets --from-literal 'mysql_password=v@ulTi$d0p3'
```

```
kubectl create configmap vault-configs --from-literal 'mysql_username=vault'
```

Create the `template` configmap:

```
kubectl create -f configmaps/template.yaml
```

Process all konfd templates in all namespaces:

```
konfd
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
    backend "mysql" {
      username = "vault"
      password = "v@ulTi$d0p3"
    }
```

### Testing with noop mode

```
konfd -onetime -noop -namespace default -configmap template
```
