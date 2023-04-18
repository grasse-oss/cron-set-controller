package controllers

import (
	"context"
	batchv1alpha1 "github.com/grasse-oss/cron-set-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					Selector: &metav1.LabelSelector{},
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
})
