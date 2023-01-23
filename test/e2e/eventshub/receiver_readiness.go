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
	"runtime/debug"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cetest "github.com/cloudevents/sdk-go/v2/test"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/pkg/apis"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/kmeta"
	"sigs.k8s.io/yaml"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/eventshub/assert"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
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
		sinkURI := apis.HTTP(fmt.Sprintf("%s.%s.svc", sinkName, ns))
		bytes, err := json.Marshal(ev)
		kevent := base64.StdEncoding.EncodeToString(bytes)
		checkError(t.Fatal, err)
		pods := kube.CoreV1().Pods(ns)
		name := feature.MakeRandomK8sName("sender")
		cmd := fmt.Sprintf("echo $K_EVENT | base64 -d "+
			"| curl --max-time 30 --trace-ascii %% --trace-time -d @- "+
			"-H \"content-type: application/cloudevents+json\" %s", sinkURI)
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
					Name:    name,
					Image:   ubi8Image,
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", cmd},
					Env: []corev1.EnvVar{{
						Name:  "K_EVENT",
						Value: kevent,
					}},
				}},
			},
		}
		_, err = pods.Create(ctx, pod, metav1.CreateOptions{})
		checkError(t.Fatal, err)
		env.Reference(kmeta.ObjectReference(pod))

		checkError(t.Error, waitForCompletion(ctx, t, pod))
		// fetch current state
		pod, err = pods.Get(ctx, pod.Name, metav1.GetOptions{})
		checkError(t.Fatal, err)
		if pod.Status.Phase != corev1.PodSucceeded {
			var logs, status []byte
			logs, err = k8s.PodLogs(ctx, pod.Name, pod.Name, pod.Namespace)
			checkError(t.Fatal, err)
			status, err = yaml.Marshal(pod.Status)
			checkError(t.Fatal, err)
			sinkLogs, err := k8s.PodLogs(ctx, sinkName, "eventshub", ns)
			checkError(t.Fatal, err)
			t.Fatalf("wanted pod to succeed, status:\n%s\n---\nLogs:\n-----\n%s\nsink logs:\n%s", status, logs, sinkLogs)
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
						Name:    name,
						Image:   ubi8Image,
						Command: []string{"/bin/sh"},
						Args:    []string{"-c", "while true; do sleep 10; done"},
					}},
				},
			},
		},
	}
	created, err := daemonSets.Create(ctx, ds, metav1.CreateOptions{})
	checkError(t.Fatal, err)
	defer func() {
		checkError(t.Fatal, daemonSets.Delete(ctx, name, metav1.DeleteOptions{}))
	}()
	env.Reference(kmeta.ObjectReference(created))
	checkError(t.Error, waitForDaemonSetReady(ctx, t, ds))
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
				t.Log(ds.Namespace, ds.Name, err)
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
				t.Log(pod.Namespace, pod.Name, err)
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

func checkError(fn func(...interface{}), err error) {
	if err != nil {
		fn("unexpected error:", err, "\n", string(debug.Stack()))
	}
}
