/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	"github.com/go-logr/logr"
	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const OwnerLabel = "grasse.io/owner"

// CronSetReconciler reconciles a CronSet object
type CronSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *CronSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.CronSet{}).
		Owns(&batchv1.CronJob{}).
		Complete(r); err != nil {
		return err
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Watches(&source.Kind{Type: &corev1.Node{}},
			handler.EnqueueRequestsFromMapFunc(func(node client.Object) []reconcile.Request {
				nodeLabels := node.GetLabels()
				r.Log.Info("Node Event", "Node", node.GetName(), "Node Labels", nodeLabels)

				var cronSetObjs batchv1alpha1.CronSetList
				_ = mgr.GetClient().List(context.TODO(), &cronSetObjs)

				var requests []reconcile.Request

			CronSetLoop:
				for _, cronSet := range cronSetObjs.Items {
					nodeSelector := cronSet.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.NodeSelector
					r.Log.Info("Check CronSet for Node Event", "CronSet name", cronSet.Name, "nodeSelector", nodeSelector)

					if len(nodeSelector) != 0 {
						for key, value := range nodeSelector {
							nodeLabel, exists := nodeLabels[key]
							if !exists || nodeLabel != value {
								r.Log.Info("CronSet's nodeSelector doesn't match event trigger node's label", "Node", node.GetName(), "CronSet", cronSet.Name)
								continue CronSetLoop
							}
						}
					}

					r.Log.Info("Add to request CronSet due to node event occured", "Node", node.GetName(), "CronSet", cronSet.Name)
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      cronSet.Name,
							Namespace: cronSet.Namespace,
						},
					})
				}

				return requests
			})).Complete(r); err != nil {
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups=batch.grasse.io,resources=cronsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.grasse.io,resources=cronsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.grasse.io,resources=cronsets/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CronSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Reconcile:", "request name", req.Name, "request namespace", req.Namespace)

	cronSet := &batchv1alpha1.CronSet{}
	if err := r.Get(ctx, req.NamespacedName, cronSet); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var nodeSelector map[string]string
	if cronSet.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.NodeSelector != nil {
		nodeSelector = cronSet.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.NodeSelector
	}
	nodeList := &corev1.NodeList{}
	nodeMap := make(map[string]bool)
	if err := r.List(ctx, nodeList, client.MatchingLabels(nodeSelector)); err != nil && errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	for _, node := range nodeList.Items {
		if err := r.applyCronJob(ctx, cronSet, node.Name); err != nil {
			return ctrl.Result{RequeueAfter: 5}, nil
		}
		nodeMap[node.Name] = true
	}

	err := r.cleanUpCronJob(ctx, cronSet.Name, nodeMap)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CronSetReconciler) applyCronJob(ctx context.Context, cronSet *batchv1alpha1.CronSet, nodeName string) error {
	cronJobName := generateCronJobName(cronSet.Name, nodeName)
	cronJobKey := metav1.ObjectMeta{
		Name:      cronJobName,
		Namespace: cronSet.Namespace,
	}
	cronJob := &batchv1.CronJob{
		ObjectMeta: cronJobKey,
	}

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, cronJob, func() error {
		updateCronJobSpec(cronJob, cronSet, nodeName)
		return controllerutil.SetControllerReference(cronSet, cronJob, r.Scheme)
	})
	if err != nil && errors.IsInvalid(err) {
		_ = r.Delete(ctx, &batchv1.CronJob{ObjectMeta: cronJobKey}, client.PropagationPolicy("Background"))
		return err
	}

	return nil
}

func (r *CronSetReconciler) cleanUpCronJob(ctx context.Context, cronSetName string, nodeMap map[string]bool) error {
	cronJobList := &batchv1.CronJobList{}
	cronSetSelector := map[string]string{OwnerLabel: cronSetName}
	if err := r.List(ctx, cronJobList, client.MatchingLabels(cronSetSelector)); err != nil && !errors.IsNotFound(err) {
		return err
	}

	for _, cronJob := range cronJobList.Items {
		if _, exist := nodeMap[cronJob.Spec.JobTemplate.Spec.Template.Spec.NodeName]; !exist {
			if err := r.Delete(ctx, &cronJob, &client.DeleteOptions{}); err != nil {
				return err
			}
		}
	}

	return nil
}

func generateCronJobName(cronSetName string, nodeName string) string {
	return strings.Join([]string{cronSetName, nodeName, "cronjob"}, "-")
}

func updateCronJobSpec(cronJob *batchv1.CronJob, cronSet *batchv1alpha1.CronSet, nodeName string) {
	cronJobSpec := cronSet.Spec.CronJobTemplate.Spec
	cronJobSpec.JobTemplate.Spec.Template.Spec.NodeName = nodeName

	cronJobLabels := cronSet.Labels
	if cronJobLabels == nil {
		cronJobLabels = make(map[string]string)
	}
	cronJobLabels[OwnerLabel] = cronSet.Name

	cronJob.ObjectMeta.Labels = cronJobLabels
	cronJob.Spec = cronJobSpec
}
