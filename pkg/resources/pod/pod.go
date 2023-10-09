/*
Copyright 2021 The Knative Authors

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

package pod

import (
	"context"
	"embed"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/manifest"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
)

//go:embed *.yaml
var yaml embed.FS

func Install(name string, image string, opts ...manifest.CfgFn) feature.StepFn {
	cfg := map[string]interface{}{
		"name": name,
	}

	for _, fn := range opts {
		fn(cfg)
	}

	return func(ctx context.Context, t feature.T) {
		if err := registerImage(ctx, image); err != nil {
			t.Fatal(err)
		}

		if ic := environment.GetIstioConfig(ctx); ic.Enabled {
			manifest.WithIstioPodAnnotations(cfg)
		}

		manifest.PodSecurityCfgFn(ctx, t)(cfg)

		if _, err := manifest.InstallYamlFS(ctx, yaml, cfg); err != nil {
			t.Fatal(err)
		}
	}
}

func registerImage(ctx context.Context, image string) error {
	reg := environment.RegisterPackage(image)
	_, err := reg(ctx, environment.FromContext(ctx))
	return err
}

func WaitForCompleted(name string, timing ...time.Duration) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		interval, timeout := k8s.PollTimings(ctx, timing)

		err := wait.PollImmediate(interval, timeout, func() (bool, error) {
			client := kubeclient.Get(ctx)
			namespace := environment.FromContext(ctx).Namespace()

			pod, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					t.Logf("pod %s not found", name)
					// keep polling
					return false, nil
				}
				return false, err
			}

			return pod.Status.Phase == corev1.PodSucceeded, nil
		})

		if err != nil {
			t.Fatal(err)
		}
	}

}
