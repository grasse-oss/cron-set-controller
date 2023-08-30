# Getting started
TBD

## Installing with Helm
TBD

## Setting for developers
The operator is based on Operator SDK with kubebuilder. so you can run this app using `main.go`.

The operator uses unit tests and [Kuttl](https://kuttl.dev/) for e2e test to make sure that the operator is working.

### Local development using make run
#### Unit test
You can run unit test using `make test`

#### Local development
We use [Kind](https://kind.sigs.k8s.io/) cluster to test this operator in local env.
You need to install kind and make new cluster.
Also there is a make target `kind-cluster` which will start a new instance of a kind cluster.

After that, you need to deploy this controller. that is also you can run make `deploy`. But as you know we should load the image from your local to the cluster.
If you want to do all things, just use the command as below.
```bash
make deploy IMG="cronset-controller:v0.0.1" KIND_CLUSTER="kind-cluster"
```
`
### E2e tests using Kuttl
As mentioned above we use Kuttl to run e2s tests for the operator. we normally run Kuttl on Kind
So please set up kind cluster and load image before executing the command as below.

The `make kuttl`