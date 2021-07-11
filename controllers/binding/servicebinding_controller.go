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

package binding

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	bindingv1beta1 "github.com/kubepreset/kubepreset/apis/binding/v1beta1"
)

// ServiceBindingRoot points to the environment variable in the container
// which is used as the volume mount path.  In the absence of this
// environment variable, `/bindings` is used as the volume mount path.
// Refer: https://github.com/k8s-service-bindings/spec#reconciler-implementation
const ServiceBindingRoot = "SERVICE_BINDING_ROOT"

// Status of a ProvisionedService
// The name will be a reference to a secret
type Status struct {
	Binding corev1.LocalObjectReference `json:"binding"`
}

// ProvisionedService represents the duck-type which could any backing service
type ProvisionedService struct {
	Status Status `json:"status"`
}

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	mountPathDir       string
	volumeNamePrefix   string
	volumeName         string
	unstructuredVolume map[string]interface{}
}

// AppNameSelectorInvariantErr represents the error when the application
// is specified through both name and label selector
type AppNameSelectorInvariantErr struct {
	Name     string
	Selector *metav1.LabelSelector
}

// Error implements the built-in error interface
func (err AppNameSelectorInvariantErr) Error() string {
	return fmt.Sprintf("Name: %v, Selector: %v", err.Name, *err.Selector)
}

// ContainersWithEnvsOrVolumeMountsErr represents the error when the ClusterApplicationResourceMapping
// is specified with Containers list and Envs or VolumeMounts
type ContainersWithEnvsOrVolumeMountsErr struct {
	Containers   []string
	Envs         []string
	VolumeMounts []string
}

// Error implements the built-in error interface
func (err ContainersWithEnvsOrVolumeMountsErr) Error() string {
	return fmt.Sprintf("Containers: %v, Envs: %v, VolumeMount: %v", err.Containers, err.Envs, err.VolumeMounts)
}

// +kubebuilder:rbac:groups=service.binding,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.binding,resources=servicebindings/status,verbs=get;update;patch

// Reconcile based on changes in the ServiceBinding CR or Provisioned Service Secret
func (r *ServiceBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("servicebinding", req.NamespacedName)

	log.V(0).Info("starting reconciliation")

	var sb bindingv1beta1.ServiceBinding

	log.V(2).Info("retrieving ServiceBinding object", "ServiceBinding", sb)
	if err := r.Get(ctx, req.NamespacedName, &sb); err != nil {
		log.Error(err, "unable to retrieve ServiceBinding")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.V(2).Info("ServiceBinding object retrieved", "ServiceBinding", sb)

	// name of the custom finalizer
	finalizerName := "binding.kubepreset.dev/finalizer"
	var conditionStatus bindingv1beta1.ConditionStatus
	var reason string

	var secretName string
	if sb.Status.Binding != nil && sb.Status.Binding.Name != "" {
		secretName = sb.Status.Binding.Name
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if sb.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have the finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering the finalizer.
		if !containsString(sb.GetFinalizers(), finalizerName) {
			controllerutil.AddFinalizer(&sb, finalizerName)
			if err := r.Update(ctx, &sb); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(sb.GetFinalizers(), finalizerName) {
			// finalizer is present, so lets handle any external dependency
			// TODO: Unbind existing bindings
			applications, result, err := r.getApplication(ctx, log, req, sb, secretName)
			if err != nil {
				return result, err
			}
			result, err = r.unbindApplications(ctx, log, req, sb, applications...)
			if err != nil {
				return result, err
			}

			// remove the finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&sb, finalizerName)
			if err := r.Update(ctx, &sb); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	var secretLookupKey client.ObjectKey

	if sb.Spec.Service.Kind == "Secret" && sb.Spec.Service.APIVersion == "v1" {
		secretLookupKey = client.ObjectKey{Name: sb.Spec.Service.Name, Namespace: req.NamespacedName.Namespace}
	} else {
		backingServiceCRLookupKey := client.ObjectKey{Name: sb.Spec.Service.Name, Namespace: req.NamespacedName.Namespace}

		backingServiceCR := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       sb.Spec.Service.Kind,
				"apiVersion": sb.Spec.Service.APIVersion,
				"metadata": map[string]interface{}{
					"name": sb.Spec.Service.Name,
				},
			},
		}

		log.V(2).Info("retrieving the backing service object", "backingServiceCR", backingServiceCR)
		if err := r.Get(ctx, backingServiceCRLookupKey, backingServiceCR); err != nil {
			reason = "unable to retrieve the backing service"
			log.Error(err, reason)
			conditionStatus = "False"
			if sb.Status.Binding != nil && sb.Status.Binding.Name != "" {
				// TODO: Unbind existing bindings
				applications, result, err := r.getApplication(ctx, log, req, sb, sb.Status.Binding.Name)
				if err != nil {
					return result, err
				}
				result, err = r.unbindApplications(ctx, log, req, sb, applications...)
				if err != nil {
					return result, err
				}

				if _, err = r.setStatus(ctx, log, sb.Status.Binding.Name, sb, conditionStatus, reason); err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, nil
			} else {
				if _, err = r.setStatus(ctx, log, "", sb, conditionStatus, reason); err != nil {
					return ctrl.Result{}, err
				}
				// Requeue with a time interval is required as the Secret name is not available to reconcile
				// In future, probably watching for provisioned services based on label can be introduced
				// or a webhook can detect provisioned service resource change and trigger reconciliation
				return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
			}
		}
		log.V(1).Info("backing service object retrieved", "backingServiceCR", backingServiceCR)

		ps := &ProvisionedService{}

		log.V(2).Info("mapping backing service with the provisioned service")
		if err := mergo.Map(ps, backingServiceCR.Object, mergo.WithOverride); err != nil {
			log.Error(err, "unable to map backing service with the provisioned service")
			return ctrl.Result{}, err
		}
		log.V(1).Info("completed mapping backing service with the provisioned service", "ProvisionedService", ps)

		secretLookupKey = client.ObjectKey{Name: ps.Status.Binding.Name, Namespace: req.NamespacedName.Namespace}
	}
	psSecret := &corev1.Secret{}

	log.V(1).Info("retrieving the Secret object")
	if err := r.Get(ctx, secretLookupKey, psSecret); err != nil {
		reason = "unable to retrieve the Secret object"
		log.Error(err, reason, "Secret Lookup Key", secretLookupKey, "Secret", psSecret)
		// TODO: Unbind existing bindings
		applications, result, err := r.getApplication(ctx, log, req, sb, psSecret.GetName())
		if err != nil {
			return result, err
		}
		result, err = r.unbindApplications(ctx, log, req, sb, applications...)
		if err != nil {
			return result, err
		}

		conditionStatus = "False"
		return r.setStatus(ctx, log, secretLookupKey.Name, sb, conditionStatus, reason)
	}
	log.V(2).Info("the secret object retrieved", "Secret", psSecret)

	if sb.Spec.Application.Name != "" && sb.Spec.Application.Selector != nil {
		err := AppNameSelectorInvariantErr{
			Name:     sb.Spec.Application.Name,
			Selector: sb.Spec.Application.Selector}
		log.Error(err, "Both application name and selector cannot be used together")
		conditionStatus = "False"
		reason = "application name and selector cannot be used together"
		return r.setStatus(ctx, log, psSecret.Name, sb, conditionStatus, reason)
	}

	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: sb.Name}}
	cm.Namespace = sb.Namespace
	cm.Labels = sb.DeepCopy().GetLabels()
	cm.Data = map[string]string{}
	if sb.Spec.Type != "" {
		cm.Data["type"] = sb.Spec.Type
	}
	if sb.Spec.Provider != "" {
		cm.Data["provider"] = sb.Spec.Provider
	}

	cm.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sb.GetObjectMeta(), sb.GroupVersionKind())}
	log.V(1).Info("Creating ConfigMap resource for binding", "ConfigMap", cm)
	if err := r.Create(ctx, cm); err != nil {
		log.Error(err, "unable to create ConfigMap resource")
		if err := r.Delete(ctx, cm); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
	log.V(1).Info("ConfigMap created", "ConfigMap", cm)

	volumeNamePrefix := sb.Name
	if len(volumeNamePrefix) > 56 {
		volumeNamePrefix = volumeNamePrefix[:56]
	}
	r.volumeName = volumeNamePrefix + "-" + psSecret.GetResourceVersion()
	r.mountPathDir = sb.Name
	if sb.Spec.Name != "" {
		r.mountPathDir = sb.Spec.Name
	}
	sp := &corev1.SecretProjection{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: psSecret.Name,
		}}
	cmp := &corev1.ConfigMapProjection{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: cm.Name,
		}}
	volumeProjection := &corev1.Volume{
		Name: r.volumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{{Secret: sp}, {ConfigMap: cmp}},
			},
		},
	}

	log.V(2).Info("converting the volumeProjection to an unstructured object", "Volume", volumeProjection)
	var err error
	r.unstructuredVolume, err = runtime.DefaultUnstructuredConverter.ToUnstructured(volumeProjection)
	if err != nil {
		log.Error(err, "unable to convert volumeProjection to an unstructured object")
		return ctrl.Result{}, err
	}

	applications, result, err := r.getApplication(ctx, log, req, sb, secretName)
	if err != nil {
		return result, err
	}
	return r.bindApplications(ctx, log, req, sb, psSecret, applications...)
}

type errorList []error

func (el errorList) Error() string {
	msg := ""
	for _, e := range el {
		msg += e.Error() + " "
	}
	return msg
}

func (r *ServiceBindingReconciler) getApplication(ctx context.Context, log logr.Logger, req ctrl.Request,
	sb bindingv1beta1.ServiceBinding, secretName string) ([]unstructured.Unstructured, ctrl.Result, error) {
	var applications []unstructured.Unstructured
	var conditionStatus bindingv1beta1.ConditionStatus
	var reason string

	if sb.Spec.Application.Name != "" {
		applicationLookupKey := client.ObjectKey{Name: sb.Spec.Application.Name, Namespace: req.NamespacedName.Namespace}

		application := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       sb.Spec.Application.Kind,
				"apiVersion": sb.Spec.Application.APIVersion,
				"metadata": map[string]interface{}{
					"name": sb.Spec.Application.Name,
				},
			},
		}

		log.V(2).Info("retrieving the application object", "Application", application)
		if err := r.Get(ctx, applicationLookupKey, application); err != nil {
			reason = "unable to retrieve application"
			log.Error(err, reason)
			conditionStatus = "False"
			result, err := r.setStatus(ctx, log, secretName, sb, conditionStatus, reason)
			return []unstructured.Unstructured{}, result, err
		}
		log.V(1).Info("application object retrieved", "Application", application)
		applications = append(applications, *application)
	}

	if sb.Spec.Application.Selector != nil {
		applicationList := &unstructured.UnstructuredList{
			Object: map[string]interface{}{
				"kind":       sb.Spec.Application.Kind,
				"apiVersion": sb.Spec.Application.APIVersion,
				"metadata": map[string]interface{}{
					"name": sb.Spec.Application.Name,
				},
			},
		}

		log.V(2).Info("retrieving the application objects", "Application", applicationList)
		opts := &client.ListOptions{
			LabelSelector: labels.Set(sb.Spec.Application.Selector.MatchLabels).AsSelector(),
		}
		if err := r.List(ctx, applicationList, opts); err != nil {
			reason = "unable to retrieve application"
			log.Error(err, reason)
			conditionStatus = "False"
			result, err := r.setStatus(ctx, log, secretName, sb, conditionStatus, reason)
			return []unstructured.Unstructured{}, result, err
		}
		log.V(1).Info("application objects retrieved", "Application", applicationList)
		applications = append(applications, applicationList.Items...)
	}
	return applications, ctrl.Result{}, nil
}

func (r *ServiceBindingReconciler) unbindApplications(ctx context.Context, log logr.Logger, req ctrl.Request,
	sb bindingv1beta1.ServiceBinding, applications ...unstructured.Unstructured) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *ServiceBindingReconciler) bindApplications(ctx context.Context, log logr.Logger, req ctrl.Request,
	sb bindingv1beta1.ServiceBinding, psSecret *corev1.Secret, applications ...unstructured.Unstructured) (ctrl.Result, error) {

	gvk := applications[0].GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	rm, err := r.RESTMapper().RESTMapping(gk, gvk.Version)
	if err != nil {
		log.Error(err, "unable to determine the RESTMapping")
		return ctrl.Result{}, err
	}

	armObj := &bindingv1beta1.ClusterApplicationResourceMapping{}

	log.V(2).Info("retrieving the ClusterApplicationResourceMapping objects", "ClusterApplicationResourceMapping", armObj)
	armLookupKey := client.ObjectKey{Name: rm.Resource.Resource + "." + gvk.Group, Namespace: req.NamespacedName.Namespace}
	armExists := true
	if err := r.Get(ctx, armLookupKey, armObj); err != nil {
		log.Error(err, "unable to retrieve ClusterApplicationResourceMapping")
		armExists = false
	}
	log.V(1).Info("ClusterApplicationResourceMapping objects retrieved", "ClusterApplicationResourceMapping", armObj)

	var el errorList
	for _, application := range applications {
		containersPaths := [][]string{}
		volumesPath := []string{"spec", "template", "spec", "volumes"}
		if armExists {
			for _, ver := range armObj.Spec.Versions {
				if ver.Version == gvk.Version || ver.Version == "*" {
					if len(ver.Containers) > 0 && (len(ver.VolumeMounts) > 0 || len(ver.Envs) > 0) {
						err := ContainersWithEnvsOrVolumeMountsErr{
							Containers:   ver.Containers,
							Envs:         ver.Envs,
							VolumeMounts: ver.VolumeMounts}
						reason := "a combination of envs and volumeMounts is mutually exclusive with containers"
						log.Error(err, reason)
						var conditionStatus bindingv1beta1.ConditionStatus = "False"
						return r.setStatus(ctx, log, psSecret.Name, sb, conditionStatus, reason)
					}
					for _, container := range ver.Containers {
						containersPaths = append(containersPaths, strings.Split(container[1:], "."))
					}
					volumesPath = strings.Split(ver.Volumes[1:], ".")
					break
				}
			}
		} else {
			containersPaths = append(containersPaths,
				[]string{"spec", "template", "spec", "containers"},
				[]string{"spec", "template", "spec", "initContainers"},
			)
		}
		log.V(2).Info("referencing the volume in an unstructured object")
		volumes, found, err := unstructured.NestedSlice(application.Object, volumesPath...)
		if !found {
			log.V(2).Info("volumes not found in the application object")
		}
		if err != nil {
			log.Error(err, "unable to reference the volumes in the application object")
			return ctrl.Result{}, err
		}
		log.V(2).Info("Volumes values", "volumes", volumes)

		volumeFound := false

		for i, volume := range volumes {
			log.V(2).Info("Volume", "volume", volume)
			if strings.HasPrefix(volume.(map[string]interface{})["name"].(string), r.volumeNamePrefix) {
				volumes[i] = r.unstructuredVolume
				volumeFound = true
			}
		}

		if !volumeFound {
			volumes = append(volumes, r.unstructuredVolume)
		}
		log.V(2).Info("setting the updated volumes into the application using the unstructured object")
		if err := unstructured.SetNestedSlice(application.Object, volumes, volumesPath...); err != nil {
			return ctrl.Result{}, err
		}
		log.V(1).Info("application object after setting the update volume", "Application", application)

		for _, containersPath := range containersPaths {
			log.V(2).Info("referencing containers in an unstructured object")
			containers, found, err := unstructured.NestedSlice(application.Object, containersPath...)
			if !found {
				e := &field.Error{Type: field.ErrorTypeRequired, Field: strings.Join(containersPath, "."), Detail: "empty containers"}
				log.V(0).Info("containers not found in the application object", "error", e)
			}
			if err != nil {
				log.Error(err, "unable to referenc containers in the application object")
				return ctrl.Result{}, err
			}

		CONTAINERS_OUTER:
			for i := range containers {
				container := &containers[i]
				log.V(2).Info("updating container", "container", container)
				c := &corev1.Container{}
				u := (*container).(map[string]interface{})
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, c); err != nil {
					return ctrl.Result{}, err
				}

				if len(sb.Spec.Application.Containers) > 0 {
					found := false
					count := 0
					for _, v := range sb.Spec.Application.Containers {
						log.V(2).Info("container", "container value", v, "name", c.Name)
						if v.StrVal == c.Name {
							break
						}
						found = true
						count++
					}
					if found && len(sb.Spec.Application.Containers) == count {
						continue CONTAINERS_OUTER
					}

				}

				for _, e := range sb.Spec.Env {
					c.Env = append(c.Env, corev1.EnvVar{
						Name:  e.Name,
						Value: string(psSecret.Data[e.Key]),
					})

				}
				mountPath := ""
				for _, e := range c.Env {
					if e.Name == ServiceBindingRoot {
						mountPath = path.Join(e.Value, r.mountPathDir)
						break
					}
				}

				if mountPath == "" {
					mountPath = path.Join("/bindings", r.mountPathDir)
					c.Env = append(c.Env, corev1.EnvVar{
						Name:  ServiceBindingRoot,
						Value: "/bindings",
					})
				}

				volumeMount := corev1.VolumeMount{
					Name:      r.volumeName,
					MountPath: mountPath,
					ReadOnly:  true,
				}

				volumeMountFound := false
				for j, vm := range c.VolumeMounts {
					if strings.HasPrefix(vm.Name, r.volumeNamePrefix) {
						c.VolumeMounts[j] = volumeMount
						volumeMountFound = true
						break
					}
				}

				if !volumeMountFound {
					c.VolumeMounts = append(c.VolumeMounts, volumeMount)
				}

				nu, err := runtime.DefaultUnstructuredConverter.ToUnstructured(c)
				if err != nil {
					return ctrl.Result{}, err
				}

				containers[i] = nu
			}

			log.V(1).Info("updated cntainer with volume and volume mounts", "containers", containers)

			log.V(2).Info("setting the updated containers into the application using the unstructured object")
			if err := unstructured.SetNestedSlice(application.Object, containers, containersPath...); err != nil {
				return ctrl.Result{}, err
			}
			log.V(1).Info("application object after setting the updated containers", "Application", application)
		}

		var conditionStatus bindingv1beta1.ConditionStatus
		var reason string

		conditionStatus = "True"

		log.V(2).Info("updating the application with updated volumes and volumeMounts")
		if err := r.Update(ctx, &application); err != nil {
			log.Error(err, "unable to update the application", "application", application)
			conditionStatus = "False"
			reason = "application update failed"
		}

		_, err = r.setStatus(ctx, log, psSecret.Name, sb, conditionStatus, reason)
		if err != nil {
			el = append(el, err)
		}
	}
	if len(el) > 0 {
		return ctrl.Result{}, el
	}
	return ctrl.Result{}, nil
}

func (r *ServiceBindingReconciler) setStatus(ctx context.Context, log logr.Logger, secretName string,
	sb bindingv1beta1.ServiceBinding, conditionStatus bindingv1beta1.ConditionStatus, reason string) (ctrl.Result, error) {

	sb.Status.Binding = &corev1.LocalObjectReference{Name: secretName}

	conditionFound := false
	for k, cond := range sb.Status.Conditions {
		if cond.Type == bindingv1beta1.ConditionReady {
			cond.Status = conditionStatus
			sb.Status.Conditions[k] = cond
			conditionFound = true
		}
	}

	if !conditionFound {
		c := bindingv1beta1.Condition{
			LastTransitionTime: metav1.NewTime(time.Now()),
			Type:               bindingv1beta1.ConditionReady,
			Status:             conditionStatus,
			Reason:             reason,
		}
		sb.Status.Conditions = append(sb.Status.Conditions, c)
	}

	log.V(2).Info("updating the service binding status")
	if err := r.Status().Update(ctx, &sb); err != nil {
		log.Error(err, "unable to update the service binding", "ServiceBinding", sb)
		return ctrl.Result{}, err
	}
	log.V(1).Info("service binding status updated", "ServiceBinding", sb)

	return ctrl.Result{}, nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// SetupWithManager setup controller with manager
func (r *ServiceBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapSecretToServiceBinding := func(a client.Object) []reconcile.Request {
		serviceBindings := &bindingv1beta1.ServiceBindingList{}
		opts := &client.ListOptions{}
		if err := r.List(context.Background(), serviceBindings, opts); err != nil {
			return []reconcile.Request{}
		}
		reply := make([]reconcile.Request, 0, len(serviceBindings.Items))
		for _, sb := range serviceBindings.Items {
			if sb.Status.Binding != nil && sb.Status.Binding.Name == a.GetName() {
				reply = append(reply, reconcile.Request{NamespacedName: types.NamespacedName{
					Namespace: sb.Namespace,
					Name:      sb.Name,
				}})
			}
		}
		return reply
	}

	genPred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&bindingv1beta1.ServiceBinding{}).
		Watches(&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(mapSecretToServiceBinding)).
		WithEventFilter(genPred).
		Complete(r)
}
