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

package v1

import (
	"context"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-batch-v1-job,mutating=true,failurePolicy=fail,sideEffects=None,groups=batch,resources=jobs,verbs=create;update,versions=v1,name=vjob.kb.io,admissionReviewVersions={v1}

type JobValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (jb *JobValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	job := &batchv1.Job{}

	err := jb.decoder.Decode(req, job)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	key := "binding.kubepreset.dev/modified"
	anno, found := job.Annotations[key]
	if !found {
		return admission.Denied(fmt.Sprintf("missing annotation %s", key))
	}
	if anno != "true" {
		return admission.Denied(fmt.Sprintf("annotation %s did not have value %q", key, "foo"))
	}
	job.Annotations["binding.kubepreset.dev/modified"] = "true"

	return admission.Allowed("")
}

// JobValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (jb *JobValidator) InjectDecoder(d *admission.Decoder) error {
	jb.decoder = d
	return nil
}
