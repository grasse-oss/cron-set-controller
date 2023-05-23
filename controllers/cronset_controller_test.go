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

var nodeA = &corev1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-node-a",
	},
}

var nodeB = &corev1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-node-b",
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

var nodeACronJobKey = types.NamespacedName{
	Name:      generateCronJobName(CronSetName, nodeA.Name),
	Namespace: CronSetNamespace,
}

var nodeBCronJobKey = types.NamespacedName{
	Name:      generateCronJobName(CronSetName, nodeB.Name),
	Namespace: CronSetNamespace,
}

var expectedOwnerRefs = []metav1.OwnerReference{{
	APIVersion: "batch.grasse.io/v1alpha1", Kind: "CronSet", Name: CronSetName,
	Controller: &trueVal, BlockOwnerDeletion: &trueVal,
}}

func TestCronSetController(t *testing.T) {
	ctx := context.Background()
	scheme, err := batchv1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	require.NoError(t, corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, batchv1.SchemeBuilder.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(nodeA).WithObjects(cronSet).Build()
	reconciler := CronSetReconciler{
		fakeClient,
		ctrl.Log.WithName("controllers").WithName("CronSet"),
		scheme,
	}

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("cronJob should be created on node a", func(t *testing.T) {
		createdCronJob := &batchv1.CronJob{}
		err = fakeClient.Get(ctx, nodeACronJobKey, createdCronJob)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCronJob)
		assert.Equal(t, expectedOwnerRefs, createdCronJob.OwnerReferences)
	})

	createdNode := &corev1.Node{}
	err = fakeClient.Create(ctx, nodeB)
	assert.NoError(t, err)
	err = fakeClient.Get(ctx, types.NamespacedName{Name: nodeB.Name}, createdNode)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdNode)

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("cronJob should also be created for newly created node b", func(t *testing.T) {
		createdCronJob := &batchv1.CronJob{}
		err = fakeClient.Get(ctx, nodeBCronJobKey, createdCronJob)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCronJob)
		assert.Equal(t, expectedOwnerRefs, createdCronJob.OwnerReferences)
	})

	t.Run("spec of cronset should be updated without issue", func(t *testing.T) {
		createdCronSet := &batchv1alpha1.CronSet{}
		err = fakeClient.Get(ctx, cronSetKey, createdCronSet)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCronSet)

		createdCronSet.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name = "test-container-2"
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
		assert.Equal(t, 2, len(updatedCronJobs.Items))

		for _, createdCronJob := range updatedCronJobs.Items {
			assert.Equal(t, "test-container-2", createdCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name)
		}
	})

	err = fakeClient.Delete(ctx, nodeB)
	assert.NoError(t, err)

	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(t, err)

	t.Run("when node b is removed, cronjob b should also be deleted", func(t *testing.T) {
		deletedCronJob := &batchv1.CronJob{}
		err = fakeClient.Get(ctx, nodeBCronJobKey, deletedCronJob)
		assert.Equal(t, true, errors.IsNotFound(err))
	})
}
