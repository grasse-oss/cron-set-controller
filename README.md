# cron-set-controller

**cron-set-controller** is a type of composite controller that provides a way to deploy CronJob resources across a specific number of nodes.
It aims to support the following features:

- Schedule Pods in a CronJob-like manner.
- Select nodes to deploy in a manner similar to a DaemonSet.
- Provide a mechanism to delay node termination when scheduled Pods exist.

### prerequisites

- kubernetes: v1.21 or later
