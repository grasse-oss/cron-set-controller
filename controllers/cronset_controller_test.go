package controllers

import (
	"context"
	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CronSet controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		CronSetName      = "test-cronset"
		CronSetNamespace = "default"
	)

	Context("When ...", func() {
		It("Should ...", func() {
			By("By creating a new CronSet")
			ctx := context.Background()
			cronJob := &batchv1alpha1.CronSet{
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
									// For simplicity, we only fill out the required fields.
									Template: v1.PodTemplateSpec{
										Spec: v1.PodSpec{
											// For simplicity, we only fill out the required fields.
											Containers: []v1.Container{
												{
													Name:  "test-container",
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
			Expect(k8sClient.Create(ctx, cronJob)).Should(Succeed())
		})
	})

	Context("Node Event", func() {
		It("Should support watching for node events", func() {
			ctx := context.Background()

			By("Creating a new node")
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
				Spec: corev1.NodeSpec{
					ProviderID: "grasse:///test/1",
				},
			}
			Expect(k8sClient.Create(ctx, node)).Should(Succeed())

			By("Checking that the mapping function has been called")
			nodeLookupKey := types.NamespacedName{Name: node.Name}
			createdNode := &corev1.Node{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, nodeLookupKey, createdNode)
				if err != nil {
					return false
				}
				return true
			}).Should(BeTrue())
		})
	})
})
