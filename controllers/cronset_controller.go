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
	"fmt"
	"github.com/go-logr/logr"
	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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
	r.Log.Info("Reconcile")
	fmt.Println("Reconcile -> {}", req)

	// TODO(user): your logic here
	obj := &batchv1alpha1.CronSet{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// for testing
	r.Log.Info(obj.GetObjectKind().GroupVersionKind().Group)
	r.Log.Info(obj.GetObjectKind().GroupVersionKind().Version)
	r.Log.Info(obj.GetObjectKind().GroupVersionKind().Kind)

	fmt.Println(obj.GetObjectKind().GroupVersionKind().Group)
	fmt.Println(obj.GetObjectKind().GroupVersionKind().Version)
	fmt.Println(obj.GetObjectKind().GroupVersionKind().Kind)

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
