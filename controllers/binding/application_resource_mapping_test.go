/*
Copyright 2020 The KubePreset Authors.

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

package binding_test

import (
	"context"
	"time"

	custompod "github.com/kubepreset/custompod/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bindingv1beta1 "github.com/kubepreset/kubepreset/apis/binding/v1beta1"
)

var _ = Describe("Application Resource Mapping:", func() {

	const (
		timeout       = time.Second * 20
		interval      = time.Millisecond * 250
		testNamespace = "default"
		podTimeout    = time.Minute * 7
		podInterval   = time.Second * 20
	)

	Context("When creating ServiceBinding with Application Resource Mapping", func() {

		AfterEach(func() {
			ctx := context.Background()
			matchLabels := map[string]string{
				"environment": "test7",
			}

			app := &custompod.CustomPod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app7",
					Labels:    matchLabels,
					Namespace: testNamespace,
				}}

			err := k8sClient.Delete(ctx, app, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			customPodLookupKey := types.NamespacedName{Name: "app7", Namespace: testNamespace}
			deletedCustomPod := &custompod.CustomPod{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, customPodLookupKey, deletedCustomPod)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret7",
					Namespace: testNamespace,
				}}
			err = k8sClient.Delete(ctx, secret, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			secretLookupKey := types.NamespacedName{Name: "secret7", Namespace: testNamespace}
			deletedSecret := &corev1.Secret{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, deletedSecret)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			sb := &bindingv1beta1.ServiceBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ServiceBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb7",
					Namespace: testNamespace,
				}}

			err = k8sClient.Delete(ctx, sb, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb7", Namespace: testNamespace}
			deletedServiceBinding := &bindingv1beta1.ServiceBinding{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceBindingLookupKey, deletedServiceBinding)
				return err != nil
			}, timeout, interval).Should(BeTrue())

		})

		It("should update the application based on the mapping and ServiceBinding status conditions for type `Ready` with value `True`", func() {
			ctx := context.Background()

			By("Creating Secret")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret7",
					Namespace: testNamespace,
				},
				StringData: map[string]string{
					"type":     "custom",
					"provider": "backingservice",
					"username": "guest",
					"password": "password",
				},
			}
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

			matchLabels := map[string]string{
				"environment": "test7",
			}

			app := &custompod.CustomPod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app7",
					Labels:    matchLabels,
					Namespace: testNamespace,
				},
				Spec: custompod.CustomPodSpec{
					Containers: []corev1.Container{{
						Image: "ghcr.io/kubepreset/bindingdata:latest",
						Name:  "bindingdata",
					}},
				},
			}
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			arm := &bindingv1beta1.ClusterApplicationResourceMapping{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ClusterApplicationResourceMapping",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custompods.binding.kubepreset.dev",
					Namespace: testNamespace,
				},
				Spec: bindingv1beta1.ClusterApplicationResourceMappingSpec{
					Versions: []bindingv1beta1.ClusterApplicationResourceMappingVersion{{
						Version:    "v1beta1",
						Containers: []string{".spec.containers"},
						Volumes:    ".spec.volumes",
					}},
				},
			}
			Expect(k8sClient.Create(ctx, arm)).Should(Succeed())

			sb := &bindingv1beta1.ServiceBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ServiceBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb7",
					Namespace: testNamespace,
				},
				Spec: bindingv1beta1.ServiceBindingSpec{
					Application: &bindingv1beta1.Application{
						APIVersion: "binding.kubepreset.dev/v1beta1",
						Kind:       "CustomPod",
						Name:       "app7",
					},
					Service: &bindingv1beta1.Service{
						APIVersion: "v1",
						Kind:       "Secret",
						Name:       "secret7",
					},
					Env: []bindingv1beta1.Environment{
						{Name: "BACKING_SERVICE_USERNAME", Key: "username"},
						{Name: "BACKING_SERVICE_PASSWORD", Key: "password"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, sb)).Should(Succeed())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb7", Namespace: testNamespace}
			createdServiceBinding := &bindingv1beta1.ServiceBinding{}

			// Retry getting newly created ServiceBinding; the status may not be immediately reflected.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceBindingLookupKey, createdServiceBinding)
				if err != nil {
					return false
				}
				for _, condition := range createdServiceBinding.Status.Conditions {
					if condition.Type == bindingv1beta1.ConditionReady &&
						condition.Status == bindingv1beta1.ConditionTrue {
						return true
					}
				}
				return false

			}, podTimeout, podInterval).Should(BeTrue())

			Expect(len(createdServiceBinding.Status.Conditions)).To(Equal(1))
			Expect(createdServiceBinding.Status.Binding.Name).To(Equal("sb7"))

			applicationLookupKey := types.NamespacedName{Name: sb.Spec.Application.Name, Namespace: testNamespace}

			Expect(k8sClient.Get(ctx, applicationLookupKey, app)).Should(Succeed())
			Expect(len(app.Spec.Volumes)).To(Equal(1))
			Expect(app.Spec.Volumes[0].Name).To(HavePrefix("sb7-"))
			Expect(app.Spec.Volumes[0].VolumeSource.Projected.Sources[0].Secret.Name).To(Equal("secret7"))
			Expect(app.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(app.Spec.Containers[0].VolumeMounts[0].Name).To(HavePrefix("sb7-"))
			Expect(app.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/bindings/sb7"))
		})
	})
})
