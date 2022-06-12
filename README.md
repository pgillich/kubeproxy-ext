# kube-proxy extension

It injects additional information to Pods of the response (kubectl columns).
This container is a reverse proxy
to a Kubernetes API provided by kube-proxy (without authentication).

<!-- markdownlint-disable MD013 -->

## Development

### kubectl proxy

Starting a kubectl proxy for local development:

```sh
kubectl proxy --reject-methods=POST,PUT,PATCH -v5
```

Starting a kubectl proxy for remote development:

```sh
kubectl proxy --address 0.0.0.0 --accept-hosts='.*' --reject-methods=POST,PUT,PATCH -v5
```

### Updating k8s dependencies

Example:

```sh
tools/download-deps.sh v1.21.13
```

## Using

Running the service from shell:

```sh
./build/bin/kubeproxy-ext
```

Example for getting Pod list:

```sh
curl 127.0.0.1:8003/api/v1/namespaces/kubernetes-dashboard/pods
```
