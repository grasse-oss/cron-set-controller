package controllers

import (
	"context"
	"testing"

	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	CronSetName      = "test-cronset"
	CronSetNamespace = "default"
)

var trueVal = true

var node = &corev1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-node",
	},
}

var cronSet = &batchv1alpha1.CronSet{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "batch.grasse.io/v1alpha1",
		Kind:       "CronSet",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      CronSetName,
		Namespace: CronSetNamespace,
	},
	Spec: batchv1alpha1.CronSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"foo": "bar"},
		},
		CronJobTemplate: batchv1alpha1.CronJobTemplateSpec{
			Spec: batchv1.CronJobSpec{
				Schedule: "1 * * * *",
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:  "test-container-1",
										Image: "test-image",
									},
								},
								RestartPolicy: v1.RestartPolicyOnFailure,
							},
						},
					},
				},
			},
		},
	},
}

var cronSetKey = types.NamespacedName{
	Name:      CronSetName,
	Namespace: CronSetNamespace,
}

var nodeCronJobKey = types.NamespacedName{
	Name:      generateCronJobName(CronSetName, node.Name),
	Namespace: CronSetNamespace,
}

var expectedOwnerRefs = []metav1.OwnerReference{{
	APIVersion: "batch.grasse.io/v1alpha1", Kind: "CronSet", Name: CronSetName,
	Controller: &trueVal, BlockOwnerDeletion: &trueVal,
}}

func TestCreatesCronJob(t *testing.T) {
	ctx := context.Background()
	scheme, err := batchv1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, batchv1.SchemeBuilder.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).WithObjects(cronSet).Build()
	reconciler := CronSetReconciler{
		fakeClient,
		ctrl.Log.WithName("controllers").WithName("CronSet"),
		scheme,
	}

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("cronJob should be created on node", func(t *testing.T) {
		createdCronJob := &batchv1.CronJob{}
		err = fakeClient.Get(ctx, nodeCronJobKey, createdCronJob)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCronJob)
		assert.Equal(t, expectedOwnerRefs, createdCronJob.OwnerReferences)
	})
}

func TestCreatesCronJobAfterNodeAdded(t *testing.T) {
	ctx := context.Background()
	scheme, err := batchv1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, batchv1.SchemeBuilder.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).WithObjects(cronSet).Build()
	reconciler := CronSetReconciler{
		fakeClient,
		ctrl.Log.WithName("controllers").WithName("CronSet"),
		scheme,
	}

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	newNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-node",
		},
	}

	newNodeCronJobKey := types.NamespacedName{
		Name:      generateCronJobName(CronSetName, newNode.Name),
		Namespace: CronSetNamespace,
	}

	createdNode := &corev1.Node{}
	err = fakeClient.Create(ctx, newNode)
	assert.NoError(t, err)
	err = fakeClient.Get(ctx, types.NamespacedName{Name: newNode.Name}, createdNode)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdNode)

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("cronJob should also be created for newly created node", func(t *testing.T) {
		createdCronJob := &batchv1.CronJob{}
		err = fakeClient.Get(ctx, newNodeCronJobKey, createdCronJob)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCronJob)
		assert.Equal(t, expectedOwnerRefs, createdCronJob.OwnerReferences)
	})
}

func TestUpdatesCronJobWhenCronSetUpdated(t *testing.T) {
	ctx := context.Background()
	scheme, err := batchv1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, batchv1.SchemeBuilder.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).WithObjects(cronSet).Build()
	reconciler := CronSetReconciler{
		fakeClient,
		ctrl.Log.WithName("controllers").WithName("CronSet"),
		scheme,
	}

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	newContainerName := "new-test-container"

	t.Run("spec of cronset should be updated without issue", func(t *testing.T) {
		createdCronSet := &batchv1alpha1.CronSet{}
		err = fakeClient.Get(ctx, cronSetKey, createdCronSet)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCronSet)

		createdCronSet.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name = newContainerName
		err = fakeClient.Update(ctx, createdCronSet)
		assert.NoError(t, err)
	})

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("if cronset changes, cronjobs should also change", func(t *testing.T) {
		updatedCronJobs := &batchv1.CronJobList{}
		cronSetSelector := map[string]string{OwnerLabel: cronSet.Name}
		err = fakeClient.List(ctx, updatedCronJobs, client.MatchingLabels(cronSetSelector))
		assert.NoError(t, err)
		assert.NotEmpty(t, updatedCronJobs)
		assert.Equal(t, 1, len(updatedCronJobs.Items))

		for _, createdCronJob := range updatedCronJobs.Items {
			assert.Equal(t, newContainerName, createdCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name)
		}
	})
}

func TestDeletesCronJobWhenNodeIsRemoved(t *testing.T) {
	ctx := context.Background()
	scheme, err := batchv1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, batchv1.SchemeBuilder.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).WithObjects(cronSet).Build()
	reconciler := CronSetReconciler{
		fakeClient,
		ctrl.Log.WithName("controllers").WithName("CronSet"),
		scheme,
	}

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("cronJob should be created on node", func(t *testing.T) {
		createdCronJob := &batchv1.CronJob{}
		err = fakeClient.Get(ctx, nodeCronJobKey, createdCronJob)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCronJob)
		assert.Equal(t, expectedOwnerRefs, createdCronJob.OwnerReferences)
	})

	err = fakeClient.Delete(ctx, node)
	assert.NoError(t, err)

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("when node is removed, cronjob should also be deleted", func(t *testing.T) {
		deletedCronJob := &batchv1.CronJob{}
		err = fakeClient.Get(ctx, nodeCronJobKey, deletedCronJob)
		assert.Equal(t, true, errors.IsNotFound(err))
	})
}
