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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/kmeta"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/manifest"
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

func IsRunning(name string) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		ns := environment.FromContext(ctx).Namespace()
		podClient := kubeclient.Get(ctx).CoreV1().Pods(ns)
		p := podClient
		interval, timeout := environment.PollTimings(ctx, nil)
		err := wait.PollImmediate(interval, timeout, func() (bool, error) {
			p, err := p.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					t.Log("pod", "namespace", ns, "name", name, err)
					// keep polling
					return false, nil
				}
				return true, err
			}
			isRunning := podRunning(p)

			if !isRunning {
				t.Logf("Pod %s/%s is not running...", ns, name)
			}

			return isRunning, nil
		})
		if err != nil {
			sb := strings.Builder{}
			if p, err := podClient.Get(ctx, name, metav1.GetOptions{}); err != nil {
				sb.WriteString(err.Error())
				sb.WriteString("\n")
			} else {
				sb.WriteString("Pod: ")
				podJson, _ := json.MarshalIndent(p, "", "  ")
				sb.WriteString(string(podJson))
				sb.WriteString("\n")
				for _, c := range p.Spec.Containers {
					if b, err := Logs(ctx, name, c.Name, environment.FromContext(ctx).Namespace()); err != nil {
						sb.WriteString(err.Error())
					} else {
						sb.Write(b)
					}
					sb.WriteString("\n")
				}
			}
			t.Fatalf("Failed while waiting for pod %s running: %+v\n%s\n", name, errors.WithStack(err), sb.String())
		}
	}
}

func registerImage(ctx context.Context, image string) error {
	reg := environment.RegisterPackage(image)
	_, err := reg(ctx, environment.FromContext(ctx))
	return err
}

// Logs returns Pod logs for given Pod and Container in the namespace
func Logs(ctx context.Context, podName, containerName, namespace string) ([]byte, error) {
	podClient := kubeclient.Get(ctx).CoreV1().Pods(namespace)
	podList, err := podClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for i := range podList.Items {
		// Pods are big, so avoid copying.
		pod := &podList.Items[i]
		if strings.Contains(pod.Name, podName) {
			result := podClient.GetLogs(pod.Name, &corev1.PodLogOptions{
				Container: containerName,
			}).Do(ctx)
			return result.Raw()
		}
	}
	return nil, fmt.Errorf("could not find logs for %s/%s:%s", namespace, podName, containerName)
}

func Reference(namespace string, name string) (corev1.ObjectReference, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: namespace,
		},
	}
	scheme := runtime.NewScheme()
	err := corev1.SchemeBuilder.AddToScheme(scheme)
	if err != nil {
		return corev1.ObjectReference{}, errors.WithStack(err)
	}
	kinds, _, err := scheme.ObjectKinds(pod)
	if err != nil {
		return corev1.ObjectReference{}, errors.WithStack(err)
	}
	if !(len(kinds) > 0) {
		return corev1.ObjectReference{}, errors.New("want len(kinds) > 0")
	}
	kind := kinds[0]
	pod.APIVersion, pod.Kind = kind.ToAPIVersionAndKind()
	return kmeta.ObjectReference(pod), nil
}

// podRunning will check the status conditions of the pod and return true if it's Running.
func podRunning(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded
}
