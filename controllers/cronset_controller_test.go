package controllers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
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
		Name:        "test-node",
		Labels:      map[string]string{"foo": "bar"},
		Annotations: map[string]string{"xyz": "baz"},
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
		CronJobTemplate: batchv1alpha1.CronJobTemplateSpec{
			Spec: batchv1.CronJobSpec{
				Schedule: "1 * * * *",
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "test-container-1",
										Image: "test-image",
									},
								},
								RestartPolicy: corev1.RestartPolicyOnFailure,
								NodeSelector:  map[string]string{"foo": "bar"},
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

type CronSetSuite struct {
	suite.Suite
	reconciler CronSetReconciler
	fakeClient client.Client
}

func (s *CronSetSuite) SetupTest() {
	ctx = context.Background()
	scheme, err := batchv1alpha1.SchemeBuilder.Build()
	require.NoError(s.T(), err)
	require.NoError(s.T(), corev1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(s.T(), batchv1.SchemeBuilder.AddToScheme(scheme))

	s.fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(node).WithObjects(cronSet).Build()

	s.reconciler = CronSetReconciler{
		s.fakeClient,
		ctrl.Log.WithName("controllers").WithName("CronSet"),
		scheme,
	}
}

func TestCronSetSuite(t *testing.T) {
	suite.Run(t, new(CronSetSuite))
}

/*
	TC Function Format =>
	Test<Event Category>_<Event>_<Result>
*/

func (s *CronSetSuite) TestCronSetEvent_Create_CreateCronJob() {
	s.Run("When reconcile after creating a CronSet object", func() {
		_, err := s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("Should create a CronJob object into the relevant nodes", func() {
			createdCronJob := &batchv1.CronJob{}
			err := s.fakeClient.Get(ctx, nodeCronJobKey, createdCronJob)
			assert.NoError(s.T(), err)
			assert.NotEmpty(s.T(), createdCronJob)
			assert.Equal(s.T(), expectedOwnerRefs, createdCronJob.OwnerReferences)
		})
	})

	s.Run("When reconcile after creating a CronSet object with NodeIdentificationKey Env", func() {
		os.Setenv(NodeIdentificationKey, "xyz")
		_, err := s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("Should create a CronJob with a name that contains the value of the NodeIdentificationKey annotation.", func() {
			identificationKey := os.Getenv(NodeIdentificationKey)
			createdCronJob := &batchv1.CronJob{}
			key := types.NamespacedName{
				Name:      generateCronJobName(CronSetName, node.Annotations[identificationKey]),
				Namespace: CronSetNamespace,
			}
			err := s.fakeClient.Get(ctx, key, createdCronJob)
			assert.NoError(s.T(), err)
			assert.NotEmpty(s.T(), createdCronJob)
			assert.Equal(s.T(), expectedOwnerRefs, createdCronJob.OwnerReferences)
		})

		os.Unsetenv(NodeIdentificationKey)
	})

	s.Run("When reconcile after creating a CronSet object with wrong NodeIdentificationKey Env", func() {
		os.Setenv(NodeIdentificationKey, "abc")
		_, err := s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("Should include node's name in the name of the generated cronjob.", func() {
			createdCronJob := &batchv1.CronJob{}
			err := s.fakeClient.Get(ctx, nodeCronJobKey, createdCronJob)
			assert.NoError(s.T(), err)
			assert.NotEmpty(s.T(), createdCronJob)
			assert.Equal(s.T(), expectedOwnerRefs, createdCronJob.OwnerReferences)
		})

		os.Unsetenv(NodeIdentificationKey)
	})
}

func (s *CronSetSuite) TestCronSetEvent_Update_UpdateCronJob() {
	s.Run("When updating a CronSet object", func() {
		createdCronSet := &batchv1alpha1.CronSet{}
		err := s.fakeClient.Get(ctx, cronSetKey, createdCronSet)
		assert.NoError(s.T(), err)
		assert.NotEmpty(s.T(), createdCronSet)

		newContainerName := "new-test-container"
		createdCronSet.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name = newContainerName
		err = s.fakeClient.Update(ctx, createdCronSet)
		assert.NoError(s.T(), err)

		_, err = s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("Should update a CronJob object", func() {
			updatedCronJobs := &batchv1.CronJobList{}
			cronSetSelector := map[string]string{OwnerLabel: cronSet.Name}
			err = s.fakeClient.List(ctx, updatedCronJobs, client.MatchingLabels(cronSetSelector))
			assert.NoError(s.T(), err)
			assert.NotEmpty(s.T(), updatedCronJobs)
			assert.Equal(s.T(), 1, len(updatedCronJobs.Items))

			for _, createdCronJob := range updatedCronJobs.Items {
				assert.Equal(s.T(), newContainerName, createdCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Name)
			}
		})
	})
}

func (s *CronSetSuite) TestCronSetEvent_UpdateNodeSelector_DeleteCronJob() {
	createdCronSet := &batchv1alpha1.CronSet{}
	_ = s.fakeClient.Get(ctx, cronSetKey, createdCronSet)
	s.Run("When updating a CronSet nodeSelector using that is different with node label", func() {
		createdCronSet.Spec.CronJobTemplate.Spec.JobTemplate.Spec.Template.Spec.NodeSelector = map[string]string{"foo": "bar1"}

		err := s.fakeClient.Update(ctx, createdCronSet)
		assert.NoError(s.T(), err)

		_, err = s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("Should delete a CronJob object", func() {
			deletedCronJob := &batchv1.CronJob{}
			err = s.fakeClient.Get(ctx, nodeCronJobKey, deletedCronJob)
			assert.Equal(s.T(), true, errors.IsNotFound(err))
		})
	})
}

func (s *CronSetSuite) TestNodeEvent_Create_CreateCronJob() {
	newNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "new-node",
			Labels: map[string]string{"foo": "bar"},
		},
	}

	newNodeCronJobKey := types.NamespacedName{
		Name:      generateCronJobName(CronSetName, newNode.Name),
		Namespace: CronSetNamespace,
	}

	s.Run("When creating a Node", func() {
		createdNode := &corev1.Node{}
		err := s.fakeClient.Create(ctx, newNode)
		assert.NoError(s.T(), err)
		err = s.fakeClient.Get(ctx, types.NamespacedName{Name: newNode.Name}, createdNode)
		assert.NoError(s.T(), err)
		assert.NotEmpty(s.T(), createdNode)

		_, err = s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("should create a CronJob object into the Node that is relevant with CronSet", func() {
			createdCronJob := &batchv1.CronJob{}
			err = s.fakeClient.Get(ctx, newNodeCronJobKey, createdCronJob)
			assert.NoError(s.T(), err)
			assert.NotEmpty(s.T(), createdCronJob)
			assert.Equal(s.T(), expectedOwnerRefs, createdCronJob.OwnerReferences)
		})
	})
}

func (s *CronSetSuite) TestNodeEvent_Delete_RemoveCronJob() {
	_, err := s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(s.T(), err)

	createdCronJob := &batchv1.CronJob{}
	err = s.fakeClient.Get(ctx, nodeCronJobKey, createdCronJob)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), createdCronJob)
	assert.Equal(s.T(), expectedOwnerRefs, createdCronJob.OwnerReferences)

	s.Run("When deleting a Node which applied a CronJob object", func() {
		err = s.fakeClient.Delete(ctx, node)
		assert.NoError(s.T(), err)

		_, err = s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("Should delete the CronJob object", func() {
			deletedCronJob := &batchv1.CronJob{}
			err = s.fakeClient.Get(ctx, nodeCronJobKey, deletedCronJob)
			assert.Equal(s.T(), true, errors.IsNotFound(err))
		})
	})
}

func (s *CronSetSuite) TestNodeEvent_UpdateLabel_RemoveCronJob() {
	_, err := s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
	assert.NoError(s.T(), err)

	createdCronJob := &batchv1.CronJob{}
	err = s.fakeClient.Get(ctx, nodeCronJobKey, createdCronJob)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), createdCronJob)

	createdNode := &corev1.Node{}
	err = s.fakeClient.Get(ctx, types.NamespacedName{Name: node.Name}, createdNode)
	assert.NoError(s.T(), err)

	s.Run("When updating a node label using that is different with CronSet nodeSelector", func() {
		createdNode.Labels = map[string]string{"foo": "bar1"}

		err := s.fakeClient.Update(ctx, createdNode)
		assert.NoError(s.T(), err)

		_, err = s.reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: cronSetKey})
		assert.NoError(s.T(), err)

		s.Run("Should delete the CronJob object", func() {
			deletedCronJob := &batchv1.CronJob{}
			err = s.fakeClient.Get(ctx, nodeCronJobKey, deletedCronJob)
			assert.Equal(s.T(), true, errors.IsNotFound(err))
		})
	})
}
