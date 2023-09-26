/*
 * Copyright 2023 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package serviceaccount

import (
	"context"
	"embed"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/pkg/client/injection/kube/client"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/manifest"
)

//go:embed *.yaml
var yaml embed.FS
var WithLabels = manifest.WithLabels
var WithAnnotations = manifest.WithAnnotations

func Install(name string, opts ...manifest.CfgFn) feature.StepFn {
	cfg := map[string]interface{}{
		"name": name,
	}

	for _, fn := range opts {
		fn(cfg)
	}

	return func(ctx context.Context, t feature.T) {
		if _, err := manifest.InstallYamlFS(ctx, yaml, cfg); err != nil {
			t.Fatal(err)
		}
	}
}

// WaitForBeingAvailable waits until the given service account was created by
// some 3rd party
func WaitForBeingAvailable(name string, timing ...time.Duration) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		env := environment.FromContext(ctx)
		interval, timeout := k8s.PollTimings(ctx, timing)
		var lastErr error

		err := wait.PollImmediate(interval, timeout, func() (bool, error) {
			_, err := client.Get(ctx).
				CoreV1().
				ServiceAccounts(env.Namespace()).
				Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				lastErr = err
				return false, nil
			}

			return true, nil
		})
		if err != nil {
			t.Fatalf("failed to get service account %s/%s: %v. Last error: %v", name, env.Namespace(), err, lastErr)
		}
	}
}
