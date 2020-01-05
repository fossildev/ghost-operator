# Ghost Operator

The Ghost Operator is an implementation of a [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) using [Operator SDK](https://github.com/operator-framework/operator-sdk) for [Ghost](https://ghost.org/) headless CMS for professional publishing.

## Project Status

This project is currently work-in-progress and in Alpha, so it may not be production ready.

## Getting Started

> We assume you already have a running cluster

To install the operator, run:

```bash
kubectl create -f https://raw.githubusercontent.com/fossildev/ghost-operator/master/deploy/crds/ghost.fossil.or.id_ghostapps_crd.yaml
kubectl create -f https://raw.githubusercontent.com/fossildev/ghost-operator/master/deploy/service_account.yaml
kubectl create -f https://raw.githubusercontent.com/fossildev/ghost-operator/master/deploy/role.yaml
kubectl create -f https://raw.githubusercontent.com/fossildev/ghost-operator/master/deploy/role_binding.yaml
kubectl create -f https://raw.githubusercontent.com/fossildev/ghost-operator/master/deploy/operator.yaml
```

Once the `ghost-operator` deployment is ready, create a GhostApp instance, like:

```bash
kubectl apply -f - << EOF
apiVersion: ghost.fossil.or.id/v1alpha1
kind: GhostApp
metadata:
  name: example-ghostapp
spec:
  replicas: 1
  image: ghost:3
  config:
    url: http://ghost.example.com
    database:
      client: sqlite3
      connection:
        filename: /var/lib/ghost/content/data/ghost.db
  persistent:
    enabled: true
    size: 10Gi
  ingress:
    enabled: true
    hosts:
    - "ghost.example.com"
    - "www.ghost.example.com"
EOF
```

In this example, the Ghost App is available at http://ghost.example.com and Ghost Admin at http://ghost.example.com/ghost/

## Contributions

We hope you'll get involved! Read our [Contributors' Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache-2.0 License - see the [LICENSE](LICENSE) for details