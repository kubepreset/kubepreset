/*
Copyright 2021 The KubePreset Authors.

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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bindingv1beta1 "github.com/kubepreset/kubepreset/apis/binding/v1beta1"
)

var _ = Describe("Watch Direct Secret Reference:", func() {

	const (
		timeout       = time.Second * 20
		interval      = time.Millisecond * 250
		testNamespace = "default"
		podTimeout    = time.Minute * 7
		podInterval   = time.Second * 20
	)

	Context("When creating ServiceBinding with Direct Secret Reference after the ServiceBinding", func() {

		AfterEach(func() {
			ctx := context.Background()
			matchLabels := map[string]string{
				"environment": "test1",
			}

			app := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app1",
					Labels:    matchLabels,
					Namespace: testNamespace,
				}}

			err := k8sClient.Delete(ctx, app, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			deploymentLookupKey := types.NamespacedName{Name: "app1", Namespace: testNamespace}
			deletedDeployment := &appsv1.Deployment{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, deploymentLookupKey, deletedDeployment)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
					Namespace: testNamespace,
				}}
			err = k8sClient.Delete(ctx, secret, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			secretLookupKey := types.NamespacedName{Name: "secret1", Namespace: testNamespace}
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
					Name:      "sb1",
					Namespace: testNamespace,
				}}

			err = k8sClient.Delete(ctx, sb, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb1", Namespace: testNamespace}
			deletedServiceBinding := &bindingv1beta1.ServiceBinding{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceBindingLookupKey, deletedServiceBinding)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			podList := &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test1"})
				if err != nil {
					return false
				}
				if len(podList.Items) > 0 {
					return false
				}
				return true
			}, podTimeout, podInterval).Should(BeTrue())

		})

		It("should update the ServiceBinding status conditions for type `Ready` with value `True`", func() {
			ctx := context.Background()

			matchLabels := map[string]string{
				"environment": "test1",
			}

			app := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app1",
					Labels:    matchLabels,
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: matchLabels,
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: matchLabels,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Image: "ghcr.io/kubepreset/bindingdata:latest",
								Name:  "bindingdata",
							}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			podList := &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test1"})
				if err != nil {
					return false
				}
				if len(podList.Items) > 0 {
					ready := []bool{}
					for _, p := range podList.Items {
						found := false
						for _, condition := range p.Status.Conditions {
							if condition.Type == corev1.PodReady &&
								condition.Status == corev1.ConditionTrue {
								ready = append(ready, true)
								found = true
								break
							}
						}
						if !found {
							ready = append(ready, false)
						}
					}
					for _, v := range ready {
						if v == false {
							return false
						}
					}
					return true
				}
				return false
			}, podTimeout, podInterval).Should(BeTrue())

			sb := &bindingv1beta1.ServiceBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ServiceBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: testNamespace,
				},
				Spec: bindingv1beta1.ServiceBindingSpec{
					Application: &bindingv1beta1.Application{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "app1",
					},
					Service: &bindingv1beta1.Service{
						APIVersion: "v1",
						Kind:       "Secret",
						Name:       "secret1",
					},
					Env: []bindingv1beta1.Environment{
						{Name: "BACKING_SERVICE_USERNAME", Key: "username"},
						{Name: "BACKING_SERVICE_PASSWORD", Key: "password"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, sb)).Should(Succeed())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb1", Namespace: testNamespace}
			createdServiceBinding := &bindingv1beta1.ServiceBinding{}

			By("Creating Secret")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret1",
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
			Expect(createdServiceBinding.Status.Binding.Name).To(Equal("secret1"))

			applicationLookupKey := types.NamespacedName{Name: sb.Spec.Application.Name, Namespace: testNamespace}

			Expect(k8sClient.Get(ctx, applicationLookupKey, app)).Should(Succeed())
			Expect(len(app.Spec.Template.Spec.Volumes)).To(Equal(1))
			Expect(app.Spec.Template.Spec.Volumes[0].Name).To(HavePrefix("sb1-"))
			Expect(app.Spec.Template.Spec.Volumes[0].VolumeSource.Projected.Sources[0].Secret.Name).To(Equal("secret1"))
			Expect(app.Spec.Template.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Template.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Template.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(app.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(HavePrefix("sb1-"))
			Expect(app.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/bindings/sb1"))

			podList = &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test1"})
				if err != nil {
					return false
				}
				if len(podList.Items) > 0 {
					ready := []bool{}
					for _, p := range podList.Items {
						found := false
						for _, condition := range p.Status.Conditions {
							if condition.Type == corev1.PodReady &&
								condition.Status == corev1.ConditionTrue {
								ready = append(ready, true)
								found = true
								break
							}
						}
						if !found {
							ready = append(ready, false)
						}
					}
					for _, v := range ready {
						if v == false {
							return false
						}
					}
					return true
				}
				return false
			}, podTimeout, podInterval).Should(BeTrue())

			podList2 := &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList2, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test1"})
				return err == nil
			}, podTimeout, podInterval).Should(BeTrue())

			Expect(podList2.Items[0].Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			found := false
			for _, vm := range podList2.Items[0].Spec.Containers[0].VolumeMounts {
				if vm.MountPath == "/bindings/sb1" {
					found = true
					Expect(vm.Name).To(HavePrefix("sb1-"))
					Expect(vm.ReadOnly).To(Equal(true))
				}
			}
			Expect(found).To(Equal(true))
		})
	})

})
