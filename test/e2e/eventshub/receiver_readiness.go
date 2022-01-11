//go:build e2e
// +build e2e

/*
Copyright 2022 The Knative Authors

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

package eventshub

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cetest "github.com/cloudevents/sdk-go/v2/test"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/kmeta"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/eventshub/assert"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"sigs.k8s.io/yaml"
)

const ubi8Image = "registry.access.redhat.com/ubi8/ubi"

// ReceiverReadiness checks the readiness of the Events Hub receiver.
func ReceiverReadiness() *feature.Feature {
	f := feature.NewFeature()
	sinkName := feature.MakeRandomK8sName("sink")
	ev := cetest.FullEvent()
	ev.SetID(feature.MakeRandomK8sName("test-event"))

	f.Setup("Pre cache sender image", precacheSenderImage)
	f.Setup("Deploy sink", eventshub.Install(sinkName, eventshub.StartReceiver))

	f.Requirement("Send event", sendEvent(ev, sinkName))

	f.Stable("Event").Must("received", receiveEvent(ev, sinkName))

	return f
}

func sendEvent(ev cloudevents.Event, sinkName string) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		kube := kubeclient.Get(ctx)
		env := environment.FromContext(ctx)
		ns := env.Namespace()
		sinkURI := fmt.Sprintf("http://%s.%s.svc.cluster.local", sinkName, ns)
		bytes, err := json.Marshal(ev)
		kevent := base64.StdEncoding.EncodeToString(bytes)
		errorIsNil(t, err)
		pods := kube.CoreV1().Pods(ns)
		name := feature.MakeRandomK8sName("sender")
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
				Labels: map[string]string{
					"event-id": ev.ID(),
				},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{{
					Name:    "sender",
					Image:   ubi8Image,
					Command: []string{"/bin/sh"},
					Args: []string{"-c", "echo $K_EVENT | base64 -d " +
						"| curl -XPOST -d @- " +
						"-H \"content-type: application/cloudevents+json\" " +
						sinkURI},
					Env: []corev1.EnvVar{{
						Name:  "K_EVENT",
						Value: kevent,
					}},
				}},
			},
		}
		_, err = pods.Create(ctx, pod, metav1.CreateOptions{})
		errorIsNil(t, err)
		env.Reference(kmeta.ObjectReference(pod))

		errorIsNil(t, waitForCompletion(ctx, t, pod))
		// fetch current state
		pod, err = pods.Get(ctx, pod.Name, metav1.GetOptions{})
		errorIsNil(t, err)
		if pod.Status.Phase != corev1.PodSucceeded {
			logs, err := k8s.PodLogs(ctx, pod.Name, "sender", pod.Namespace)
			errorIsNil(t, err)
			status, err := yaml.Marshal(pod.Status)
			errorIsNil(t, err)
			t.Fatalf("wanted pod to success, got: \n%s\n\nLogs: %s", status, string(logs))
		}
	}
}

func precacheSenderImage(ctx context.Context, t feature.T) {
	kube := kubeclient.Get(ctx)
	env := environment.FromContext(ctx)
	ns := env.Namespace()
	// using daemon set to make sure image is cached on every node
	daemonSets := kube.AppsV1().DaemonSets(ns)
	name := feature.MakeRandomK8sName("precache-sender")
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: ns,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "sender",
						Image:   ubi8Image,
						Command: []string{"/bin/sh"},
						Args:    []string{"-c", "while true; do sleep 10; done"},
					}},
				},
			},
		},
	}
	_, err := daemonSets.Create(ctx, ds, metav1.CreateOptions{})
	errorIsNil(t, err)
	env.Reference(kmeta.ObjectReference(ds))
	errorIsNil(t, waitForDaemonSetReady(ctx, t, ds))
	errorIsNil(t, daemonSets.Delete(ctx, name, metav1.DeleteOptions{}))
}

func receiveEvent(ev cloudevents.Event, sinkName string) feature.StepFn {
	return assert.OnStore(sinkName).
		MatchEvent(cetest.HasId(ev.ID())).
		Exact(1)
}

func waitForDaemonSetReady(ctx context.Context, t feature.T, ds *appsv1.DaemonSet, timing ...time.Duration) error {
	interval, timeout := k8s.PollTimings(ctx, timing)
	kube := kubeclient.Get(ctx)
	daemonSets := kube.AppsV1().DaemonSets(ds.Namespace)

	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		current, err := daemonSets.Get(ctx, ds.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Log(ds.Namespace, ds.Name, "not found", err)
				// keep polling
				return false, nil
			}
			return false, err
		}

		t.Log(ds.Namespace, ds.Name,
			"daemon set desired", current.Status.DesiredNumberScheduled,
			"ready", current.Status.NumberReady)
		return current.Status.NumberReady >= 1 &&
			current.Status.DesiredNumberScheduled == current.Status.NumberReady, nil
	})
}

func waitForCompletion(ctx context.Context, t feature.T, pod *corev1.Pod, timing ...time.Duration) error {
	interval, timeout := k8s.PollTimings(ctx, timing)
	kube := kubeclient.Get(ctx)
	pods := kube.CoreV1().Pods(pod.Namespace)

	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		current, err := pods.Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Log(pod.Namespace, pod.Name, "not found", err)
				// keep polling
				return false, nil
			}
			return false, err
		}

		t.Log(pod.Namespace, pod.Name, "pod status phase", current.Status.Phase)
		return current.Status.Phase == corev1.PodSucceeded ||
			current.Status.Phase == corev1.PodFailed, nil
	})
}

func errorIsNil(t feature.T, err error) {
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
}
