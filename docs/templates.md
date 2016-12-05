# Templates


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
    default_lease_ttl = {{configmap "vault" "default_lease_ttl"}}
    max_lease_ttl = {{configmap "vault" "max_lease_ttl"}}
    listener "tcp" {
        address = "0.0.0.0:8200"
        tls_cert_file = "/etc/tls/server.pem"
        tls_key_file = "/etc/tls/server.key"
    }
    backend "mysql" {
        username = "{{configmap "vault" "mysql.username"}}"
        password = "{{secret "vault" "mysql.password"}}"
        address = "{{configmap "vault" "mysql.address"}}"
        database = "{{configmap "vault" "mysql.database"}}"
        table = "{{configmap "vault" "mysql.table"}}"
        tls_ca_file = "/etc/tls/mysql-ca.pem"
    }
```
