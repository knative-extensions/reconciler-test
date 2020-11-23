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

package k8s

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
)

// IsReady returns a reusable feature.StepFn to assert if a resource is ready.
func IsReady(gvr schema.GroupVersionResource, name string, interval, timeout time.Duration) feature.StepFn {
	return func(ctx context.Context, t *testing.T) {
		env := environment.FromContext(ctx)
		if err := WaitForResourceReady(ctx, env.Namespace(), name, gvr, interval, timeout); err != nil {
			t.Error(gvr, "did not become ready,", err)
		}
	}
}
