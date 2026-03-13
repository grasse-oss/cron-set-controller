# API Overview
## Architecture
![cron-set-controller-architecture.png](pictures%2Farchitecture.png)

The Cron Set Operator extends Kubernetes with Custom Resources, which basically define cronjob spec.
The controller creates Kubernetes [cronjobs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) into all nodes. If new node is launched, the controller will reconcile that creates new cronjob into the node.

## Resource model
![cron-set-controller-resource-model.png](pictures%2Fresource-model.png)
### CronSet
A CronSet declares what CronJob to launch. The controller create cronjob into all nodes.

## Behavior
The Cron Set Controller reconciles 'CronSet' in the following manner:
1. the controller create 'Kind=CronJob' resources based on the template provided by 'CronSet.spec' into all nodes.
2. the controller ensures that the CronJobs stay in all healthy nodes.

