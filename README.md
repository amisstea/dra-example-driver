# Space DRA Driver (Proof of Concept)

This repository contains a proof of concept for a DRA (Dynamic Resource Allocation) driver that provisions namespaces on the same cluster as the driver.
It's forked from the [dra-example-driver](https://github.com/kubernetes-sigs/dra-example-driver) project.

The intention is to demonstrate how *any* resource, not just devices/hardware, can be dynamically provisioned
with the necessary artifacts made available securely to running containers (through env variables, mounts, etc.)

## Quickstart and Demo

The driver itself provides access to namespaces created as requested, and this demo
walks through the process of building and installing the driver followed by
running a set of workloads that request these namespaces.

The procedure below has been tested and verified on Linux (Fedora).

### Prerequisites

* [GNU Make 3.81+](https://www.gnu.org/software/make/)
* [GNU Tar 1.34+](https://www.gnu.org/software/tar/)
* [docker v20.10+ (including buildx)](https://docs.docker.com/engine/install/)
* [kind v0.17.0+](https://kind.sigs.k8s.io/docs/user/quick-start/)
* [helm v3.7.0+](https://helm.sh/docs/intro/install/)
* [kubectl v1.18+](https://kubernetes.io/docs/reference/kubectl/)

### Demo

This demo requires a specific directory be present on the host:
```bash
mkdir /tmp/claim-artifacts
```

Create a `kind` cluster to run everything in:
```bash
./demo/create-cluster.sh
```

Build the image for the example resource driver:
```bash
./demo/build-driver.sh
```

And then install the example resource driver via `helm`:
```bash
helm upgrade -i \
  --create-namespace \
  --namespace dra-example-driver \
  dra-example-driver \
  deployments/helm/dra-example-driver
```

Double check the driver components have come up successfully:
```console
$ kubectl get pod -n dra-example-driver
NAME                                             READY   STATUS    RESTARTS   AGE
dra-example-driver-controller-7555d488db-nbd52   1/1     Running   0          1m
dra-example-driver-kubeletplugin-qwmbl           1/1     Running   0          1m
```


Next, deploy an example app that demonstrates how two pods can share a dynamically provisioned namespace:
```bash
kubectl apply --filename=demo/namespace-test.yaml
```

And verify that they come up successfully:
```console
$ kubectl get pods -n namespace-test
NAME   READY   STATUS    RESTARTS   AGE
pod0   1/1     Running   0          1m
pod1   1/1     Running   0          1m
```

Now take a look at the namespaces on the cluster. Take note of the namespace prefixed with `ephemeral-ns-`.
```console
$ kubectl get namespaces
NAME                 STATUS   AGE
default              Active   10m
dra-example-driver   Active   8m
ephemeral-ns-4rsv8   Active   1m
kube-node-lease      Active   10m
kube-public          Active   10m
kube-system          Active   10m
local-path-storage   Active   10m
namespace-test       Active   1m
```

Print the environment variables of the containers to see what was injected by the driver:
```console
$ kubectl exec -n namespace-test pod0 -- printenv
DRA_RESOURCE_DRIVER_NAME=space.resource.example.com
TEST_CLAIM_CLUSTER=https://cluster.todo
TEST_CLAIM_NAMESPACE=ephemeral-ns-4rsv8
TEST_CLAIM_KUBECONFIG=/var/run/claim-artifacts/e2140c9f-4fea-44a6-b688-2f52763194ab/kubeconfig
...

$ kubectl exec -n namespace-test pod1 -- printenv
DRA_RESOURCE_DRIVER_NAME=space.resource.example.com
TEST_CLAIM_CLUSTER=https://cluster.todo
TEST_CLAIM_NAMESPACE=ephemeral-ns-4rsv8
TEST_CLAIM_KUBECONFIG=/var/run/claim-artifacts/e2140c9f-4fea-44a6-b688-2f52763194ab/kubeconfig
...
```

Likewise, print the contents of the kubeconfig (not yet implemented) mounted to each container:
```console
$ kubectl exec -n namespace-test pod0 -- cat /etc/test-claim/kubeconfig        
TODO%

$ kubectl exec -n namespace-test pod1 -- cat /etc/test-claim/kubeconfig        
TODO%
```

To see namespace cleanup in action, delete the example app:
```bash
kubectl delete --filename=demo/namespace-test.yaml
```

Notice how the ephemeral namespace is gone:
```console
$ kubectl get namespaces
NAME                 STATUS   AGE
default              Active   12m
dra-example-driver   Active   10m
kube-node-lease      Active   12m
kube-public          Active   12m
kube-system          Active   12m
local-path-storage   Active   12m
```

Take a look at the driver logs for more insight into the process:
```bash
kubectl logs -n dra-example-driver -l app.kubernetes.io/name=dra-example-driver
```

Delete your cluster when you're done.
```bash
./demo/delete-cluster.sh
```

## References

For more information on the DRA Kubernetes feature and developing custom resource drivers, see the following resources:

* [Dynamic Resource Allocation in Kubernetes](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
* [KEP: 3063-dynamic-resource-allocation](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/3063-dynamic-resource-allocation)
