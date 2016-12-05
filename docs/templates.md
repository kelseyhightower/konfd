# Templates

konfd templates are [Go templates](https://golang.org/pkg/text/template) with access to two additional [template functions](https://golang.org/pkg/text/template/#hdr-Functions): `configmap` and `secrets`, which provide access to Kubernetes secrets and configmaps.

## Annotations

There are three (3) required annotations that determine where processed templates are stored:

* `konfd.io/kind` - The target resource kind (`configmap` or `secret`). 
* `konfd.io/name` - The configmap or secret name.
* `konfd.io/key`  - The key name where the processed template will be stored.

## Labels

Labels are used to filter which configmaps should be processed by konfd. By default konfd will only process configmaps with the following label:

```
konfd.io/template: "true"
```

> Use the `-configmap` flag to limit which configmaps are processed.

## Template Functions

### configmap

Returns the configmap value of the first argument (configmap name) and second argument (configmap key).

#### Usage

Create a configmap to hold configuration key/value pairs:

```
kubectl create configmap vault-configs --from-literal 'default_lease_ttl=768h'
```

Pass the configmap name as the first argument and the key name as the second argument:

```
default_lease_ttl = {{configmap "vault-configs" "default_lease_ttl"}}
```

Results:

```
default_lease_ttl = 768h
```

### secret

Returns the secret value of the first argument (secret name) and second argument (secret key).

#### Usage

Create a secret to hold the secret key/value pairs:

```
kubectl create secret generic vault-secrets \
  --from-literal 'mysql.password=v@ulTi$d0p3'
```

Pass the secret name as the first argument and the secret key name as the second argument:

```
password = "{{secret "vault-secrets" "mysql.password"}}"
```

Results:

```
password = "v@ulTi$d0p3"
```

## Example

The following template uses a mix of secrets and configmaps to generate a vault config file. The results of the `vault-template` configmap will be stored in a secret named `vault` in a key named `server.hcl`.

Ensure `konfd` is [running in the cluster](deployment-guide.md).

### Add configuration data

Create the `vault-secrets` secret:

```
kubectl create secret generic vault-secrets \
  --from-literal 'mysql.password=v@ulTi$d0p3'
```
```
secret "vault-secrets" created
```

Create the `vault-configs` configmap:

```
kubectl create configmap vault-configs \
  --from-literal 'default_lease_ttl=768h' \
  --from-literal 'max_lease_ttl=768h' \
  --from-literal 'mysql.username=vault' \
  --from-literal 'mysql.address=23.12.4.3:3306' \
  --from-literal 'mysql.database=vault' \
  --from-literal 'mysql.table=vault'
```

```
configmap "vault-configs" created
```

### Create the vault template configmap

Create the `vault-template` configmap:

```
cat vault-template.yaml
```

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: vault-template
  annotations:
    konfd.io/kind: secret
    konfd.io/name: vault
    konfd.io/key: server.hcl
  labels:
    konfd.io/template: "true"
data:
  template: |
    default_lease_ttl = {{configmap "vault-configs" "default_lease_ttl"}}
    max_lease_ttl = {{configmap "vault-configs" "max_lease_ttl"}}
    listener "tcp" {
      address = "0.0.0.0:8200"
      tls_cert_file = "/etc/tls/server.pem"
      tls_key_file = "/etc/tls/server.key"
    }
    backend "mysql" {
      username = "{{configmap "vault-configs" "mysql.username"}}"
      password = "{{secret "vault-secrets" "mysql.password"}}"
      address = "{{configmap "vault-configs" "mysql.address"}}"
      database = "{{configmap "vault-configs" "mysql.database"}}"
      table = "{{configmap "vault-configs" "mysql.table"}}"
      tls_ca_file = "/etc/tls/mysql-ca.pem"
    }
```

Submit the `vault-template` configmap to the Kubernetes API server:

```
kubectl create -f vault-template.yaml
```
```
configmap "vault-template" created
```

### Review the results

After the "vault-template" configmap is processed by `konfd` view the results:

```
kubectl get secrets vault -o yaml
```
```
apiVersion: v1
data:
  server.hcl: ZGVmYXVsdF9sZWFzZV90dGwgPSA3NjhoCm1heF9sZWFzZV90dGwgPSA3NjhoCmxpc3RlbmVyICJ0Y3AiIHsKICBhZGRyZXNzID0gIjAuMC4wLjA6ODIwMCIKICB0bHNfY2VydF9maWxlID0gIi9ldGMvdGxzL3NlcnZlci5wZW0iCiAgdGxzX2tleV9maWxlID0gIi9ldGMvdGxzL3NlcnZlci5rZXkiCn0KYmFja2VuZCAibXlzcWwiIHsKICB1c2VybmFtZSA9ICJ2YXVsdCIKICBwYXNzd29yZCA9ICJ2QHVsVGkkZDBwMyIKICBhZGRyZXNzID0gIjIzLjEyLjQuMzozMzA2IgogIGRhdGFiYXNlID0gInZhdWx0IgogIHRhYmxlID0gInZhdWx0IgogIHRsc19jYV9maWxlID0gIi9ldGMvdGxzL215c3FsLWNhLnBlbSIKfQo=
kind: Secret
metadata:
  creationTimestamp: 2016-12-05T14:24:07Z
  name: vault
  namespace: default
  resourceVersion: "331267"
  selfLink: /api/v1/namespaces/default/secrets/vault
  uid: 7b28717c-baf6-11e6-8f3a-42010a8a001a
type: Opaque
```

> Notice the server.hcl has been added to the existing vault secret.


#### Decoding Secrets

Secret values are base64 encoded. Decode the `server.hcl` to see the processed template:

```
kubectl get secrets vault -o 'go-template={{index .data "server.hcl"}}' | base64 -D -
```
```
default_lease_ttl = 768h
max_lease_ttl = 768h
listener "tcp" {
  address = "0.0.0.0:8200"
  tls_cert_file = "/etc/tls/server.pem"
  tls_key_file = "/etc/tls/server.key"
}
backend "mysql" {
  username = "vault"
  password = "v@ulTi$d0p3"
  address = "23.12.4.3:3306"
  database = "vault"
  table = "vault"
  tls_ca_file = "/etc/tls/mysql-ca.pem"
}
```
