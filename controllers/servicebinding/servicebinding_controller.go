/*
Copyright 2020 The KubePreset Authors

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

package servicebinding

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sbv1alpha2 "github.com/kubepreset/kubepreset/apis/servicebinding/v1alpha2"
)

// ServiceBindingRoot points to the environment variable in the container
// which is used as the volume mount path.  In the abscence of this
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

// Reconciler reconciles a ServiceBinding object
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var deploymentGK = schema.GroupKind{Group: "apps", Kind: "Deployment"}

// +kubebuilder:rbac:groups=service.binding,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.binding,resources=servicebindings/status,verbs=get;update;patch

// Reconcile based on changes in the ServiceBinding CR or Provisioned Service Secret
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("servicebinding", req.NamespacedName)

	log.V(2).Info("starting reconciliation")

	var sb sbv1alpha2.ServiceBinding

	log.V(1).Info("retrieving ServiceBinding object")
	if err := r.Get(ctx, req.NamespacedName, &sb); err != nil {
		log.Error(err, "unable to fetch ServiceBinding")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.V(2).Info("ServiceBinding object retrieved", "name", sb.Name, "annotations", sb.Annotations, "labels", sb.Labels)

	if sb.Status.ObservedGeneration == sb.Generation {
		return ctrl.Result{}, nil
	}

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

	log.V(1).Info("retrieving the backing service object")
	if err := r.Get(ctx, backingServiceCRLookupKey, backingServiceCR); err != nil {
		log.Error(err, "unable to fetch backing service")
		return ctrl.Result{}, err
	}
	log.V(2).Info("backing service object retrieved", "status", backingServiceCR.Object["status"])

	ps := &ProvisionedService{}

	log.V(2).Info("mapping backing service with the provisioned service")
	if err := mergo.Map(ps, backingServiceCR.Object, mergo.WithOverride); err != nil {
		return ctrl.Result{}, err
	}
	log.V(2).Info("completed mapping backing service with the provisioned service", "provisioned-service", ps)

	secretLookupKey := client.ObjectKey{Name: ps.Status.Binding.Name, Namespace: req.NamespacedName.Namespace}
	psSecret := &corev1.Secret{}

	log.V(1).Info("retrieving the secret object")
	if err := r.Get(ctx, secretLookupKey, psSecret); err != nil {
		log.Error(err, "unable to fetch backing service")
		return ctrl.Result{}, err
	}
	log.V(2).Info("the secret object retrieved", "secret-data", psSecret.Data)

	secret, err := r.createSecretForBinding(ctx, log, sb, psSecret)
	if err != nil {
		log.Error(err, "unable to create Secret resource")
		return ctrl.Result{}, err
	}

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

	log.V(1).Info("retrieving the application object")
	if err := r.Get(ctx, applicationLookupKey, application); err != nil {
		log.Error(err, "unable to fetch application")
		return ctrl.Result{}, err
	}
	log.V(2).Info("application object retrieved", "metadata", application.Object["metadata"])

	volumeNamePrefix := sb.Name
	if len(volumeNamePrefix) > 56 {
		volumeNamePrefix = volumeNamePrefix[:56]
	}
	volumeName := volumeNamePrefix + "-" + secret.GetResourceVersion()
	mountPathDir := sb.Name
	if sb.Spec.Name != "" {
		mountPathDir = sb.Spec.Name
	}
	volumeProjection := &corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Name,
			},
		},
	}

	log.V(2).Info("converting the volumeProjection to an unstructured object")
	unstructuredVolume, err := runtime.DefaultUnstructuredConverter.ToUnstructured(volumeProjection)
	if err != nil {
		log.Error(err, "unable to convert volumeProjection to an unstructured object")
		return ctrl.Result{}, err
	}

	log.V(2).Info("retrieving the existing volumes as an unstructured object")
	volumes, found, err := unstructured.NestedSlice(application.Object, "spec", "template", "spec", "volumes")
	if !found {
		log.V(2).Info("volumes not found in the application object")
	}
	if err != nil {
		log.Error(err, "locating volumes in the application object")
		return ctrl.Result{}, err
	}
	log.V(2).Info("Volumes values", "volumes", volumes)

	volumeFound := false

	for i, volume := range volumes {
		log.V(2).Info("Volume", "volume", volume)
		if strings.HasPrefix(volume.(map[string]interface{})["name"].(string), volumeNamePrefix) {
			volumes[i] = unstructuredVolume
			volumeFound = true
		}
	}

	if !volumeFound {
		volumes = append(volumes, unstructuredVolume)
	}
	log.V(2).Info("setting the updated volumes into the application using the unstructured object")
	if err := unstructured.SetNestedSlice(application.Object, volumes, "spec", "template", "spec", "volumes"); err != nil {
		return ctrl.Result{}, err
	}

	log.V(2).Info("retrieving the containers as an unstructured object")
	containers, found, err := unstructured.NestedSlice(application.Object, "spec", "template", "spec", "containers")
	if !found {
		e := &field.Error{Type: field.ErrorTypeRequired, Field: "spec.template.spec.containers", Detail: "empty containers"}
		log.Error(e, "containers not found in the application object")
		return ctrl.Result{}, apierrors.NewInvalid(deploymentGK, sb.Spec.Application.Name, field.ErrorList{e})
	}
	if err != nil {
		log.Error(err, "locating containers in the application object")
		return ctrl.Result{}, err
	}

	for i := range containers {
		container := &containers[i]
		// TODO Set log level to 2
		log.V(1).Info("updating container", "container", container)
		c := &corev1.Container{}
		u := (*container).(map[string]interface{})
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, c); err != nil {
			return ctrl.Result{}, err
		}

		mountPath := ""
		for _, e := range c.Env {
			if e.Name == ServiceBindingRoot {
				mountPath = path.Join(e.Value, mountPathDir)
				break
			}
		}

		if mountPath == "" {
			mountPath = path.Join("/bindings", mountPathDir)
			c.Env = append(c.Env, corev1.EnvVar{
				Name:  ServiceBindingRoot,
				Value: "/bindings",
			})
		}

		volumeMount := corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  true,
		}

		volumeMountFound := false
		for j, vm := range c.VolumeMounts {
			if strings.HasPrefix(vm.Name, volumeNamePrefix) {
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

	// TODO Set log level to 2
	log.V(1).Info("setting the updated containers into the application using the unstructured object")
	if err := unstructured.SetNestedSlice(application.Object, containers, "spec", "template", "spec", "containers"); err != nil {
		return ctrl.Result{}, err
	}

	sb.Status.ObservedGeneration = sb.Generation

	var s sbv1alpha2.ConditionStatus = "True"

	log.V(1).Info("updating the application with updated volumes and volumeMounts")
	if err := r.Update(ctx, application); err != nil {
		log.Error(err, "unable to update the application")
		s = "False"
	}

	return r.setStatus(ctx, log, sb, s)
}

func (r *Reconciler) setStatus(ctx context.Context, log logr.Logger,
	sb sbv1alpha2.ServiceBinding, value sbv1alpha2.ConditionStatus) (ctrl.Result, error) {

	sb.Status.Binding = &corev1.LocalObjectReference{Name: sb.Name}

	conditionFound := false
	for k, cond := range sb.Status.Conditions {
		if cond.Type == sbv1alpha2.ConditionReady {
			cond.Status = value
			sb.Status.Conditions[k] = cond
			conditionFound = true
		}
	}

	if !conditionFound {
		c := sbv1alpha2.Condition{
			LastTransitionTime: metav1.NewTime(time.Now()),
			Type:               sbv1alpha2.ConditionReady,
			Status:             value,
		}
		sb.Status.Conditions = append(sb.Status.Conditions, c)
	}

	log.V(1).Info("updating the service binding status")
	if err := r.Status().Update(ctx, &sb); err != nil {
		log.Error(err, "unable to update the service binding", "ServiceBinding", sb)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) createSecretForBinding(ctx context.Context, log logr.Logger,
	sb sbv1alpha2.ServiceBinding, sec *corev1.Secret) (*corev1.Secret, error) {
	newSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: sb.Name}}
	newSecret.Namespace = sb.Namespace
	secCopy := sec.DeepCopy()
	newSecret.Data = secCopy.Data
	l := sb.DeepCopy().GetLabels()
	if l == nil {
		l = make(map[string]string)
	}
	l["service.binding/name"] = sb.Name
	newSecret.Labels = l
	newSecret.Type = corev1.SecretType("service.binding/" + sb.Spec.Type)
	if val, ok := secCopy.Data["type"]; ok {
		newSecret.Data["type"] = val
	}
	if sb.Spec.Type != "" {
		newSecret.Data["type"] = []byte(sb.Spec.Type)
	}
	if val, ok := secCopy.Data["provider"]; ok {
		newSecret.Data["provider"] = val
	}
	if sb.Spec.Provider != "" {
		newSecret.Data["provider"] = []byte(sb.Spec.Provider)
	}
	gvk := sbv1alpha2.GroupVersion.WithKind("ServiceBinding")
	newSecret.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sb.GetObjectMeta(), gvk)}
	log.V(1).Info("Creating Secret resource for binding")
	if err := r.Create(ctx, newSecret); err != nil {
		log.Error(err, "unable to create Secret resource", "secret", newSecret)
		return nil, err
	}
	return newSecret, nil
}

// SetupWithManager setup controller with manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sbv1alpha2.ServiceBinding{}).
		Complete(r)
}
