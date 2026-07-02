package controllers

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TestNodeWatchPredicate pins the semantics of the predicate wired into the
// Node watch in SetupWithManager: node create/delete and label changes must
// pass through, while status-only updates (kubelet heartbeats, condition and
// image list refreshes) must be filtered out.
func TestNodeWatchPredicate(t *testing.T) {
	p := predicate.LabelChangedPredicate{}

	oldNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "predicate-node",
			Labels: map[string]string{"foo": "bar"},
		},
	}

	statusOnly := oldNode.DeepCopy()
	statusOnly.Status.Conditions = []corev1.NodeCondition{
		{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
	}
	statusOnly.Status.Images = []corev1.ContainerImage{
		{Names: []string{"busybox:latest"}, SizeBytes: 1234},
	}
	assert.False(t, p.Update(event.UpdateEvent{ObjectOld: oldNode, ObjectNew: statusOnly}),
		"status-only node update must not trigger reconcile")

	relabeled := oldNode.DeepCopy()
	relabeled.Labels = map[string]string{"foo": "baz"}
	assert.True(t, p.Update(event.UpdateEvent{ObjectOld: oldNode, ObjectNew: relabeled}),
		"node label change must trigger reconcile")

	assert.True(t, p.Create(event.CreateEvent{Object: oldNode}),
		"node creation must trigger reconcile")
	assert.True(t, p.Delete(event.DeleteEvent{Object: oldNode}),
		"node deletion must trigger reconcile")
}

// These specs run against envtest with the real manager started in
// suite_test.go, so node events flow through the Watches handler and the
// LabelChangedPredicate — unlike the unit tests above and in
// cronset_controller_test.go, which call Reconcile directly.
var _ = Describe("CronSet controller Node watch", func() {
	const (
		watchCronSetName = "watch-cronset"
		watchNamespace   = "default"
		matchLabelKey    = "cronset-watch-test"
	)

	var (
		testCtx = context.Background()
		cronSet *batchv1alpha1.CronSet
	)

	newWatchNode := func(name string, labels map[string]string) *corev1.Node {
		return &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		}
	}

	cronJobKeyFor := func(nodeName string) types.NamespacedName {
		return types.NamespacedName{
			Name:      generateCronJobName(watchCronSetName, nodeName),
			Namespace: watchNamespace,
		}
	}

	waitCronJobCreated := func(nodeName string) {
		Eventually(func() error {
			return k8sClient.Get(testCtx, cronJobKeyFor(nodeName), &batchv1.CronJob{})
		}, 10*time.Second, 250*time.Millisecond).Should(Succeed())
	}

	waitCronJobDeleted := func(nodeName string) {
		Eventually(func() bool {
			err := k8sClient.Get(testCtx, cronJobKeyFor(nodeName), &batchv1.CronJob{})
			return errors.IsNotFound(err)
		}, 10*time.Second, 250*time.Millisecond).Should(BeTrue())
	}

	BeforeEach(func() {
		cronSet = &batchv1alpha1.CronSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      watchCronSetName,
				Namespace: watchNamespace,
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
											{Name: "test-container", Image: "test-image"},
										},
										RestartPolicy: corev1.RestartPolicyOnFailure,
										NodeSelector:  map[string]string{matchLabelKey: "true"},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(testCtx, cronSet)).To(Succeed())
	})

	AfterEach(func() {
		nodeList := &corev1.NodeList{}
		Expect(k8sClient.List(testCtx, nodeList)).To(Succeed())
		for i := range nodeList.Items {
			_ = k8sClient.Delete(testCtx, &nodeList.Items[i])
		}

		Expect(k8sClient.Delete(testCtx, cronSet)).To(Succeed())

		// envtest runs no garbage collector, so owned CronJobs must be
		// removed explicitly to keep specs independent.
		cronJobList := &batchv1.CronJobList{}
		Expect(k8sClient.List(testCtx, cronJobList)).To(Succeed())
		for i := range cronJobList.Items {
			_ = k8sClient.Delete(testCtx, &cronJobList.Items[i])
		}
	})

	It("creates a CronJob when a matching Node is created", func() {
		node := newWatchNode("watch-node-create", map[string]string{matchLabelKey: "true"})
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())

		waitCronJobCreated(node.Name)
	})

	It("removes the CronJob when the Node's labels stop matching", func() {
		node := newWatchNode("watch-node-relabel", map[string]string{matchLabelKey: "true"})
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())
		waitCronJobCreated(node.Name)

		createdNode := &corev1.Node{}
		Expect(k8sClient.Get(testCtx, types.NamespacedName{Name: node.Name}, createdNode)).To(Succeed())
		createdNode.Labels = map[string]string{matchLabelKey: "false"}
		Expect(k8sClient.Update(testCtx, createdNode)).To(Succeed())

		waitCronJobDeleted(node.Name)
	})

	It("removes the CronJob when the Node is deleted", func() {
		node := newWatchNode("watch-node-delete", map[string]string{matchLabelKey: "true"})
		Expect(k8sClient.Create(testCtx, node)).To(Succeed())
		waitCronJobCreated(node.Name)

		Expect(k8sClient.Delete(testCtx, node)).To(Succeed())

		waitCronJobDeleted(node.Name)
	})
})
