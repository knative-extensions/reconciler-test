/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sequencestepper

import (
	corev1 "k8s.io/api/core/v1"
	"knative.dev/reconciler-test/pkg/config"

	"knative.dev/reconciler-test/pkg/framework"
	"knative.dev/reconciler-test/pkg/installer"
	"knative.dev/test-infra/pkg/helpers"
)

const packageName = "knative.dev/reconciler-test/images/sequencestepper"

var (
	Component = &sequenceStepperComponent{}
)

// Deploy creates a Sequence Stepper pod and service. Returns the name of the service
func Deploy(rc framework.ResourceContext) corev1.ObjectReference {
	image := rc.ImageName(packageName)
	name := helpers.AppendRandomString("seq-stepper")

	rc.CreateFromYAMLOrFail(installer.ExecuteTemplate(podTemplate, map[string]interface{}{"Name": name, "Image": image}))
	rc.CreateFromYAMLOrFail(installer.ExecuteTemplate(serviceTemplate, map[string]interface{}{"Name": name}))

	return corev1.ObjectReference{
		Namespace: rc.Namespace(),
		Name:      name,
	}
}

type sequenceStepperComponent struct {
}

var _ framework.Component = (*sequenceStepperComponent)(nil)

// Scope returns the component scope
func (s *sequenceStepperComponent) Scope() framework.ComponentScope {
	return framework.ComponentScopeResource
}

func (s *sequenceStepperComponent) Required(rc framework.ResourceContext, cfg config.Config) {
	installer.RegisterPackage(packageName)
}
