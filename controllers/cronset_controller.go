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
	"strings"

	"github.com/go-logr/logr"
	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const OwnerLabel = "owner"

// CronSetReconciler reconciles a CronSet object
type CronSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
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
	r.Log.Info("Reconcile -> ", req)

	obj := &batchv1alpha1.CronSet{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var nodeSelector map[string]string
	if obj.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.NodeSelector != nil {
		nodeSelector = obj.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.NodeSelector
	}
	nodes := &corev1.NodeList{}
	if err := r.Client.List(ctx, nodes, client.MatchingLabels(nodeSelector)); err != nil && errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	nodeMap := make(map[string]int)
	for i, item := range nodes.Items {
		nodeMap[item.Name] = i
	}

	cronJobs := &batchv1.CronJobList{}
	cronSetSelector := map[string]string{OwnerLabel: obj.Name}
	if err := r.Client.List(ctx, cronJobs, client.MatchingLabels(cronSetSelector)); err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	for _, cronJob := range cronJobs.Items {
		if _, exist := nodeMap[cronJob.Spec.JobTemplate.Spec.Template.Spec.NodeName]; !exist {
			if err := r.Client.Delete(ctx, &cronJob, &client.DeleteOptions{}); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	for nodeName := range nodeMap {
		cronJobName := generateCronJobName(obj.Name, nodeName)
		cronJob := batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: cronJobName, Namespace: obj.Namespace}}

		_, err := ctrl.CreateOrUpdate(ctx, r.Client, &cronJob, func() error {
			updateCronJob(&cronJob, obj, nodeName)
			return controllerutil.SetControllerReference(obj, &cronJob, r.Scheme)
		})

		if err != nil {
			if errors.IsInvalid(err) {
				objectMeta := metav1.ObjectMeta{Name: cronJob.Name, Namespace: cronJob.Namespace}
				_ = r.Client.Delete(ctx, &batchv1.Job{ObjectMeta: objectMeta}, client.PropagationPolicy("Background"))
				return reconcile.Result{RequeueAfter: 5}, nil
			}
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
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
			handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
				var cronSetObjs batchv1alpha1.CronSetList
				_ = mgr.GetClient().List(context.TODO(), &cronSetObjs)

				var requests []reconcile.Request
				for _, obj := range cronSetObjs.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      obj.Name,
							Namespace: obj.Namespace,
						},
					})
				}

				return requests
			})).Complete(r); err != nil {
		return err
	}

	return nil
}

func updateCronJob(cronJob *batchv1.CronJob, cronSet *batchv1alpha1.CronSet, nodeName string) {
	cronJobSpec := cronSet.Spec.CronJobTemplate.Spec
	cronJobSpec.JobTemplate.Spec.Template.Spec.NodeName = nodeName

	cronJobLabel := cronSet.Labels
	if cronJobLabel == nil {
		cronJobLabel = make(map[string]string)
	}
	cronJobLabel[OwnerLabel] = cronSet.Name

	cronJob.ObjectMeta.Labels = cronJobLabel
	cronJob.Spec = cronJobSpec
}

func generateCronJobName(cronSetName string, nodeName string) string {
	return strings.Join([]string{cronSetName, nodeName, "cronjob"}, "-")
}
