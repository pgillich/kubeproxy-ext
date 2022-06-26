# kube-proxy extension

It injects additional information to Pods of the response (kubectl columns).
This container is a reverse proxy
to a Kubernetes API provided by kube-proxy (without authentication).

<!-- markdownlint-disable MD013 -->

## Supported compressions

It decompresses `gzip` and `deflate` content encodings.

## Supported fields

Below fields are supported on Pods (same to the columns of `kubectl get pod -o wide`):

* `NAME` Name
* `READY` Ready
* `STATUS` Status
* `RESTARTS` Restarts
* `AGE` Age
* `IP` Ip
* `NODE` Node
* `NOMINATED_NODE` NominatedNode
* `READINESS_GATES` ReadinessGates
* `CONDITIONS` Conditions

> The `CONDITIONS` is filled only for Pod of a Job.

Example extended output:

```json
{
  "apiVersion": "v1",
  "kind": "Pod",
  "kubectl": {
    "Age": "126m",
    "Conditions": "<none>",
    "Ip": "10.244.1.12",
    "Name": "alertmanager-prometheus-stack-kube-prom-alertmanager-0",
    "Node": "demo-worker",
    "NominatedNode": "<none>",
    "ReadinessGates": "<none>",
    "Ready": "2/2",
    "Restarts": 0,
    "Status": "Running"
  },
  "metadata": {
    "name": "alertmanager-prometheus-stack-kube-prom-alertmanager-0",
    "namespace": "monitoring",
```

## Using

### Configuration

Supported environment variables:

* `LOGLEVEL` Log level, default: `debug`
* `PROXY_TARGETURL` URL to kubectl proxy, default: `http://localhost:8005`
* `PROXY_LISTENADDR` Listening address, default: `:8004`

### Local prereq

Starting a local kubectl proxy:

```sh
kubectl proxy --reject-methods=POST,PUT,PATCH -v5
```

### Run service manually

Running the service from shell:

```sh
./build/bin/kubeproxy-ext
```

### Run service in Kubernetes

Example deployment can be found here: <https://github.com/pgillich/grafana-kubernetes/blob/main/kubernetes/monitoring/kubectl-proxy-deployment.yaml>

### Testing

Example for getting Pod list:

```sh
curl 127.0.0.1:8003/api/v1/namespaces/kubernetes-dashboard/pods
```

## Development

### Remote prereq

Starting a kubectl proxy for remote development, if it's really needed:

```sh
kubectl proxy --address 0.0.0.0 --accept-hosts='.*' --reject-methods=POST,PUT,PATCH -v5
```

### Updating k8s dependencies

Example:

```sh
tools/download-deps.sh v1.21.13
```

### Debug in Kubernetes container with VSCode

Change `command` and `image` of `kubectl-ext` container of <https://github.com/pgillich/grafana-kubernetes/blob/main/kubernetes/monitoring/kubectl-proxy-deployment.yaml> with below values:

```yaml
      - name: kubectl-ext
        command:
        - sleep
        - infinity
        image: golangci/golangci-lint:v1.44.2
        ...
```

Set the active namespace to `monitoring` in Kubernetes explorer. Attach VSCode to `kubeclt-proxy` Pod, select `kubectl-ext` container. Upload, extract and open the kubeproxy-ext repo to the container.
