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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bindingv1beta1 "github.com/kubepreset/kubepreset/apis/binding/v1beta1"
)

var _ = Describe("Matching Containers:", func() {

	const (
		timeout       = time.Second * 20
		interval      = time.Millisecond * 250
		testNamespace = "default"
		podTimeout    = time.Minute * 7
		podInterval   = time.Second * 20
	)

	Context("When creating ServiceBinding without containers name list", func() {

		AfterEach(func() {
			ctx := context.Background()
			matchLabels := map[string]string{
				"environment": "test3",
			}

			app := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app3",
					Labels:    matchLabels,
					Namespace: testNamespace,
				}}

			err := k8sClient.Delete(ctx, app, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			deploymentLookupKey := types.NamespacedName{Name: "app3", Namespace: testNamespace}
			deletedDeployment := &appsv1.Deployment{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, deploymentLookupKey, deletedDeployment)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret3",
					Namespace: testNamespace,
				}}
			err = k8sClient.Delete(ctx, secret, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			secretLookupKey := types.NamespacedName{Name: "secret3", Namespace: testNamespace}
			deletedSecret := &corev1.Secret{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, deletedSecret)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myservice3",
					Namespace: testNamespace,
				}}
			err = k8sClient.Delete(ctx, svc, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			svcLookupKey := types.NamespacedName{Name: "myservice3", Namespace: testNamespace}
			deletedService := &corev1.Service{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, svcLookupKey, deletedService)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			sb := &bindingv1beta1.ServiceBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ServiceBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb3",
					Namespace: testNamespace,
				}}

			err = k8sClient.Delete(ctx, sb, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb3", Namespace: testNamespace}
			deletedServiceBinding := &bindingv1beta1.ServiceBinding{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceBindingLookupKey, deletedServiceBinding)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			podList := &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test3"})
				if err != nil {
					return false
				}
				if len(podList.Items) > 0 {
					return false
				}
				return true
			}, podTimeout, podInterval).Should(BeTrue())

		})

		It("should update all the containers including init containers", func() {
			ctx := context.Background()

			By("Creating Secret")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret3",
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
				"environment": "test3",
			}

			app := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app3",
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
							InitContainers: []corev1.Container{{
								Image:   "gcr.io/kubernetes-e2e-test-images/dnsutils:1.3",
								Name:    "dnsutils",
								Command: []string{"sh", "-c", "until nslookup myservice3; do echo waiting for myservice3; sleep 2s; done;"},
							}},
							Containers: []corev1.Container{{
								Image: "quay.io/kubepreset/bindingdata:latest",
								Name:  "bindingdata1",
								Args:  []string{"7080"},
							}, {
								Image: "quay.io/kubepreset/bindingdata:latest",
								Name:  "bindingdata2",
								Args:  []string{"7081"},
							}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			podList := &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test3"})
				if err != nil {
					return false
				}
				if len(podList.Items) > 0 {
					for _, p := range podList.Items {
						for _, status := range p.Status.InitContainerStatuses {
							now := metav1.Now()
							if status.Name == "dnsutils" &&
								status.State.Running != nil &&
								status.State.Running.StartedAt.Before(&now) {
								return true
							}
						}
					}
					return false
				}
				return false
			}, podTimeout, podInterval).Should(BeTrue())

			sb := &bindingv1beta1.ServiceBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ServiceBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb3",
					Namespace: testNamespace,
				},
				Spec: bindingv1beta1.ServiceBindingSpec{
					Application: &bindingv1beta1.Application{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "app3",
					},
					Service: &bindingv1beta1.Service{
						APIVersion: "v1",
						Kind:       "Secret",
						Name:       "secret3",
					},
					Env: []bindingv1beta1.Environment{
						{Name: "BACKING_SERVICE_USERNAME", Key: "username"},
						{Name: "BACKING_SERVICE_PASSWORD", Key: "password"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, sb)).Should(Succeed())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb3", Namespace: testNamespace}
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
			Expect(createdServiceBinding.Status.Binding.Name).To(Equal("sb3"))

			applicationLookupKey := types.NamespacedName{Name: sb.Spec.Application.Name, Namespace: testNamespace}

			Expect(k8sClient.Get(ctx, applicationLookupKey, app)).Should(Succeed())
			Expect(len(app.Spec.Template.Spec.Volumes)).To(Equal(1))
			Expect(app.Spec.Template.Spec.Volumes[0].Name).To(HavePrefix("sb3-"))
			Expect(app.Spec.Template.Spec.Volumes[0].VolumeSource.Projected.Sources[0].Secret.Name).To(Equal("secret3"))

			Expect(app.Spec.Template.Spec.InitContainers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Template.Spec.InitContainers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Template.Spec.InitContainers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(app.Spec.Template.Spec.InitContainers[0].VolumeMounts[0].Name).To(HavePrefix("sb3-"))
			Expect(app.Spec.Template.Spec.InitContainers[0].VolumeMounts[0].MountPath).To(Equal("/bindings/sb3"))

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myservice3",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"environment": "test3",
					},
					Ports: []corev1.ServicePort{{
						Protocol: "TCP",
						Port:     7080,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 7080,
						},
					}},
				},
			}
			Expect(k8sClient.Create(ctx, svc)).Should(Succeed())

			podList = &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test3"})
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

			Expect(app.Spec.Template.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Template.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Template.Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(app.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name).To(HavePrefix("sb3-"))
			Expect(app.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath).To(Equal("/bindings/sb3"))
			Expect(app.Spec.Template.Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Template.Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Template.Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(app.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name).To(HavePrefix("sb3-"))
			Expect(app.Spec.Template.Spec.Containers[1].VolumeMounts[0].MountPath).To(Equal("/bindings/sb3"))

			podList = &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test3"})
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
				err := k8sClient.List(ctx, podList2, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test3"})
				return err == nil
			}, podTimeout, podInterval).Should(BeTrue())

			Expect(podList2.Items[0].Spec.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			found := false
			for _, vm := range podList2.Items[0].Spec.Containers[0].VolumeMounts {
				if vm.MountPath == "/bindings/sb3" {
					found = true
					Expect(vm.Name).To(HavePrefix("sb3-"))
					Expect(vm.ReadOnly).To(Equal(true))
				}
			}
			Expect(found).To(Equal(true))

			Expect(podList2.Items[0].Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			found = false
			for _, vm := range podList2.Items[0].Spec.Containers[1].VolumeMounts {
				if vm.MountPath == "/bindings/sb3" {
					found = true
					Expect(vm.Name).To(HavePrefix("sb3-"))
					Expect(vm.ReadOnly).To(Equal(true))
				}
			}
			Expect(found).To(Equal(true))

		})
	})

	Context("When creating ServiceBinding with containers name list", func() {

		AfterEach(func() {
			ctx := context.Background()
			matchLabels := map[string]string{
				"environment": "test4",
			}

			app := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app4",
					Labels:    matchLabels,
					Namespace: testNamespace,
				}}

			err := k8sClient.Delete(ctx, app, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			deploymentLookupKey := types.NamespacedName{Name: "app4", Namespace: testNamespace}
			deletedDeployment := &appsv1.Deployment{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, deploymentLookupKey, deletedDeployment)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret4",
					Namespace: testNamespace,
				}}
			err = k8sClient.Delete(ctx, secret, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			secretLookupKey := types.NamespacedName{Name: "secret4", Namespace: testNamespace}
			deletedSecret := &corev1.Secret{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, deletedSecret)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myservice4",
					Namespace: testNamespace,
				}}
			err = k8sClient.Delete(ctx, svc, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			svcLookupKey := types.NamespacedName{Name: "myservice4", Namespace: testNamespace}
			deletedService := &corev1.Service{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, svcLookupKey, deletedService)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			sb := &bindingv1beta1.ServiceBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ServiceBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb4",
					Namespace: testNamespace,
				}}

			err = k8sClient.Delete(ctx, sb, client.GracePeriodSeconds(0))
			Expect(err).ShouldNot(HaveOccurred())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb4", Namespace: testNamespace}
			deletedServiceBinding := &bindingv1beta1.ServiceBinding{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, serviceBindingLookupKey, deletedServiceBinding)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			podList := &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test4"})
				if err != nil {
					return false
				}
				if len(podList.Items) > 0 {
					return false
				}
				return true
			}, podTimeout, podInterval).Should(BeTrue())

		})

		It("should update name matching containers including init containers", func() {
			ctx := context.Background()

			By("Creating Secret")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret4",
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
				"environment": "test4",
			}

			app := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "app4",
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
							InitContainers: []corev1.Container{{
								Image:   "gcr.io/kubernetes-e2e-test-images/dnsutils:1.3",
								Name:    "dnsutils",
								Command: []string{"sh", "-c", "until nslookup myservice4; do echo waiting for myservice4; sleep 2s; done;"},
							}},
							Containers: []corev1.Container{{
								Image: "quay.io/kubepreset/bindingdata:latest",
								Name:  "bindingdata1",
								Args:  []string{"7080"},
							}, {
								Image: "quay.io/kubepreset/bindingdata:latest",
								Name:  "bindingdata2",
								Args:  []string{"7081"},
							}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			podList := &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test4"})
				if err != nil {
					return false
				}
				if len(podList.Items) > 0 {
					for _, p := range podList.Items {
						for _, status := range p.Status.InitContainerStatuses {
							now := metav1.Now()
							if status.Name == "dnsutils" &&
								status.State.Running != nil &&
								status.State.Running.StartedAt.Before(&now) {
								return true
							}
						}
					}
					return false
				}
				return false
			}, podTimeout, podInterval).Should(BeTrue())

			sb := &bindingv1beta1.ServiceBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "binding.x-k8s.io/v1beta1",
					Kind:       "ServiceBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb4",
					Namespace: testNamespace,
				},
				Spec: bindingv1beta1.ServiceBindingSpec{
					Application: &bindingv1beta1.Application{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "app4",
						Containers: []intstr.IntOrString{{
							Type:   intstr.String,
							StrVal: "dnsutils",
						}, {
							Type:   intstr.Int,
							IntVal: 9,
						}, {
							Type:   intstr.String,
							StrVal: "bindingdata2",
						}},
					},
					Service: &bindingv1beta1.Service{
						APIVersion: "v1",
						Kind:       "Secret",
						Name:       "secret4",
					},
					Env: []bindingv1beta1.Environment{
						{Name: "BACKING_SERVICE_USERNAME", Key: "username"},
						{Name: "BACKING_SERVICE_PASSWORD", Key: "password"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, sb)).Should(Succeed())

			serviceBindingLookupKey := types.NamespacedName{Name: "sb4", Namespace: testNamespace}
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
			Expect(createdServiceBinding.Status.Binding.Name).To(Equal("sb4"))

			applicationLookupKey := types.NamespacedName{Name: sb.Spec.Application.Name, Namespace: testNamespace}

			Expect(k8sClient.Get(ctx, applicationLookupKey, app)).Should(Succeed())
			Expect(len(app.Spec.Template.Spec.Volumes)).To(Equal(1))
			Expect(app.Spec.Template.Spec.Volumes[0].Name).To(HavePrefix("sb4-"))
			Expect(app.Spec.Template.Spec.Volumes[0].VolumeSource.Projected.Sources[0].Secret.Name).To(Equal("secret4"))

			Expect(app.Spec.Template.Spec.InitContainers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Template.Spec.InitContainers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Template.Spec.InitContainers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(app.Spec.Template.Spec.InitContainers[0].VolumeMounts[0].Name).To(HavePrefix("sb4-"))
			Expect(app.Spec.Template.Spec.InitContainers[0].VolumeMounts[0].MountPath).To(Equal("/bindings/sb4"))

			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myservice4",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"environment": "test4",
					},
					Ports: []corev1.ServicePort{{
						Protocol: "TCP",
						Port:     7080,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 7080,
						},
					}},
				},
			}
			Expect(k8sClient.Create(ctx, svc)).Should(Succeed())

			podList = &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test4"})
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

			Expect(app.Spec.Template.Spec.Containers[0].Env).ShouldNot(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Template.Spec.Containers[0].Env).ShouldNot(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Template.Spec.Containers[0].Env).ShouldNot(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(len(app.Spec.Template.Spec.Containers[0].VolumeMounts)).To(Equal(0))

			Expect(app.Spec.Template.Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_USERNAME", Value: "guest"}))
			Expect(app.Spec.Template.Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "BACKING_SERVICE_PASSWORD", Value: "password"}))
			Expect(app.Spec.Template.Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			Expect(app.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name).To(HavePrefix("sb4-"))
			Expect(app.Spec.Template.Spec.Containers[1].VolumeMounts[0].MountPath).To(Equal("/bindings/sb4"))

			podList = &corev1.PodList{}
			Eventually(func() bool {
				err := k8sClient.List(ctx, podList, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test4"})
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
				err := k8sClient.List(ctx, podList2, client.InNamespace(testNamespace), client.MatchingLabels{"environment": "test4"})
				return err == nil
			}, podTimeout, podInterval).Should(BeTrue())

			Expect(podList2.Items[0].Spec.Containers[1].Env).Should(ContainElement(corev1.EnvVar{Name: "SERVICE_BINDING_ROOT", Value: "/bindings"}))
			found := false
			for _, vm := range podList2.Items[0].Spec.Containers[1].VolumeMounts {
				if vm.MountPath == "/bindings/sb4" {
					found = true
					Expect(vm.Name).To(HavePrefix("sb4-"))
					Expect(vm.ReadOnly).To(Equal(true))
				}
			}
			Expect(found).To(Equal(true))

		})
	})

})
