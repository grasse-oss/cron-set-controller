# Getting started
cron-set-controller runs within your Kubernetes cluster as a deployment resource. It utilizes CustomResourceDefinitions to configure cronjobs into every node through CronSet resources.

## Installing with Helm
The default install options will automatically install and manage the CRD as part of your helm release. But as you know, Helm doesn't support upgrading CRD yet. Please keep in your mind about that. There is a [helm crd document](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/).

```bash
helm pull oci://ghcr.io/grasse-oss/helm/cron-set-controller

helm install cron-set-controller \
    cron-set-controller-1.0.3.tgz
```

### Create your first CronSet
```yaml
apiVersion: batch.grasse.io/v1alpha1
kind: CronSet
metadata:
  name: cronset-sample
spec:
  cronJobTemplate:
    spec:
      schedule: "*/1 * * * *"
      jobTemplate:
        spec:
          template:
            spec:
              containers:
                - name: cronset-sample
                  image: busybox
                  command: ['/bin/sh']
```

## Uninstalling with Helm
Before continuing, ensure that cronset resources that have been created by users have been deleted. You can check for any existing resources with the following command:
```bash
kubectl get cronset --all-namespaces
```
Once all these resources have been deleted you are ready to uninstall cron-set-controller.

### Uninstalling with Helm
Uninstall the helm release using the delete command.
```bash
helm delete cron-set-controller
```

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