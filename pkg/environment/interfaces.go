/*
Copyright 2020 The Knative Authors

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

package environment

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/reconciler-test/pkg/feature"
)

type GlobalEnvironment interface {
	Environment() (context.Context, Environment)
}

type Environment interface {
	Test(ctx context.Context, t *testing.T, f *feature.Feature)

	Namespace() string

	RequirementLevel() feature.Levels
	FeatureState() feature.States

	Images() map[string]string
	TemplateConfig(base map[string]interface{}) map[string]interface{}

	Reference(ref ...corev1.ObjectReference)
	References() []corev1.ObjectReference

	Finish()
}
