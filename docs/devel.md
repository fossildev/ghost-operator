# Ghost Operator Development Guide

## Getting Started

This project is a regular [Kubernetes Operator](https://coreos.com/operators/)  built using the Operator SDK. Refer to the Operator SDK documentation to understand the basic architecture of this operator.

### Installing the Operator SDK command line tool

Follow the installation guidelines from [Operator SDK GitHub page](https://github.com/operator-framework/operator-sdk) or run `make install-sdk`.

### Developing

The first step is to get a local Kubernetes instance up and running. You can use any tools as you want eg: [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/), [docker-for-desktop](https://docs.docker.com/docker-for-mac/install/) or [kind](https://github.com/kubernetes-sigs/kind#installation-and-usage).

Most of development processes can be done with the `make` command. See available `make` command by executing 

```
make help
```

Once local kubernetes has finished starting, apply the CustomResourceDefinitions:

```
kubectl apply -f deploy/crds/ghost.fossil.or.id_ghostapps_crd.yaml
```

Then you can get the Operator running:

```
make run
```

At this point, a GhostApp instance can be installed:

```
kubectl apply -f deploy/example/ghost.fossil.or.id_v1alpha1_ghostapp_cr.yaml
kubectl get ghostapp
```

Example GhostApp status:

```
NAME               REPLICAS   PHASE     AGE
example-ghostapp   1          Running   12s
```

To remove the instance:

```
kubectl delete -f deploy/example/ghost.fossil.or.id_v1alpha1_ghostapp_cr.yaml
```

### Testing

Tests should be simple unit tests and/or end-to-end tests. For small changes, unit tests should be sufficient, but every new feature should be accompanied with end-to-end tests as well. Tests can be executed with:

```
make test
```

The whole set of end-to-end tests can be executed via:

```
make test-e2e
```

> NOTE: the command above requires you to build the Docker image and push to container registry. You can see instruction to build Docker image bellow.

Instead that, you can also run end-to-end tests locally with:

```
make test-e2e-local
```

### Build

To build Docker image of this operator can be executed with:

```
make build
```

> NOTE: by default, command above will build Docker image with tag `fossildev/ghost-operator:latest`. You can adjust the Docker image tag by overriding the variable `IMAGE_TAG`, like:

```
IMAGE_TAG=docker.io/yourusername/ghost-operator:latest make build
```

Then you can push the Docker image as usually

```
docker push docker.io/yourusername/ghost-operator:latest
```