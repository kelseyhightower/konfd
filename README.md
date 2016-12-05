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
  --from-literal 'mysql_username=vault'
```

Create the `template` configmap:

```
cat configmaps/template.yaml 
```
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: template
  annotations:
    konfd.io/kind: configmap
    konfd.io/name: vault
    konfd.io/key:  server.hcl
  labels:
    konfd.io/template: "true"
data:
  template: |
    backend "mysql" {
      username = "{{configmap "vault-configs" "mysql_username"}}"
      password = "{{secret "vault-secrets" "mysql_password"}}"
    }
```

Submit the `template` configmap:

```
kubectl create -f configmaps/template.yaml
```

### Deploy the konfd replicaset

```
kubectl create -f replicasets/konfd.yam
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

konfd can be run outside of the Kubernetes cluster by running `kubectl` in proxy mode:

```
kubectl proxy
```

Process a single template in the default namespace:

```
konfd -onetime -noop -namespace default -configmap template
```

```
{
  "apiVersion": "v1",
  "data": {
    "server.hcl": "default_lease_ttl = 768h\nbackend \"mysql\" {\n  username = \"vault\"\n  password = \"v@ulTi$d0p3\"\n  tls_ca_file = \"/etc/tls/mysql-ca.pem\"\n}\n"
  },
  "kind": "ConfigMap",
  "metadata": {
    "name": "vault",
    "namespace": "default"
  }
}
```
