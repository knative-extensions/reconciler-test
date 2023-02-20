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

package job

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/resources/pod"
)

//go:embed *.yaml
var yaml embed.FS

func Install(name string, image string, options ...manifest.CfgFn) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		if err := InstallError(ctx, t, name, image, options...); err != nil {
			t.Fatal(err)
		}
	}
}

func InstallError(ctx context.Context, t feature.T, name string, image string, options ...manifest.CfgFn) error {
	cfg := map[string]interface{}{
		"name":  name,
		"image": image,
	}

	for _, fn := range options {
		fn(cfg)
	}

	if err := registerImage(ctx, image); err != nil {
		return err
	}

	if ic := environment.GetIstioConfig(ctx); ic.Enabled {
		manifest.WithIstioPodAnnotations(cfg)
	}

	manifest.PodSecurityCfgFn(ctx, t)(cfg)

	if _, err := manifest.InstallYamlFS(ctx, yaml, cfg); err != nil {
		return err
	}

	return nil
}

// WithArgs adds arguments to container
func WithArgs(args []string) manifest.CfgFn {
	return func(m map[string]interface{}) {
		m["args"] = args
	}
}

// WithCommand adds command to container
func WithCommand(command []string) manifest.CfgFn {
	return func(m map[string]interface{}) {
		m["command"] = command
	}
}

func WithCompletions(c int32) manifest.CfgFn {
	return func(m map[string]interface{}) {
		m["completions"] = c
	}
}

func IsDone(name string, timing ...time.Duration) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		if err := WaitUntilJobDone(ctx, t, name, timing...); err != nil {
			t.Error("Job did not turn into done state", err)
		}
	}
}

func IsSucceeded(name string, timing ...time.Duration) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		if err := WaitUntilJobSucceeded(ctx, t, name, timing...); err != nil {
			t.Error("Job did not succeed", err)
		}
	}
}

func IsFailed(name string, timing ...time.Duration) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		if err := WaitUntilJobFailed(ctx, t, name, timing...); err != nil {
			t.Error("Job did not fail", err)
		}
	}
}

// WaitUntilJobDone waits until a job has finished.
// Timing is optional but if provided is [interval, timeout].
func WaitUntilJobDone(ctx context.Context, t feature.T, name string, timing ...time.Duration) error {
	return WaitForJobCondition(ctx, t, name, IsCompleteJob, timing...)
}

// WaitUntilJobSucceeded waits until a job has succeeded.
// Timing is optional but if provided is [interval, timeout].
func WaitUntilJobSucceeded(ctx context.Context, t feature.T, name string, timing ...time.Duration) error {
	return WaitForJobCondition(ctx, t, name, IsSucceededJob, timing...)
}

// WaitUntilJobFailed waits until a job has failed.
// Timing is optional but if provided is [interval, timeout].
func WaitUntilJobFailed(ctx context.Context, t feature.T, name string, timing ...time.Duration) error {
	return WaitForJobCondition(ctx, t, name, IsFailedJob, timing...)
}

func WaitForJobCondition(ctx context.Context, t feature.T, name string, isConditionFunc func(job *batchv1.Job) bool, timing ...time.Duration) error {
	interval, timeout := environment.PollTimings(ctx, timing)
	namespace := environment.FromContext(ctx).Namespace()
	kube := kubeclient.Get(ctx)
	jobs := kube.BatchV1().Jobs(namespace)
	var last *batchv1.Job

	err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		job, err := jobs.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("%s/%s job %+v", namespace, name, err)
				// keep polling
				return false, nil
			}
			return false, err
		}
		last = job

		conditionIsTrue := isConditionFunc(job)
		if !conditionIsTrue {
			status, err := json.Marshal(job.Status)
			if err != nil {
				return false, err
			}
			t.Logf("%s/%s job status %s", namespace, name, status)
		}
		return conditionIsTrue, nil
	})
	if err != nil {
		p, err := GetJobPodByJobName(ctx, name)
		if err != nil {
			return err
		}
		logs, err := pod.Logs(ctx, p.GetName(), "job-container", namespace)
		if err != nil {
			return err
		}
		status, err := json.MarshalIndent(last.Status, "", "  ")
		if err != nil {
			return err
		}
		return fmt.Errorf("job condition failed, status: \n%s\n---\nlogs:\n%s", string(status), logs)
	}

	return nil
}

// WaitForJobTerminationMessage waits for a job to end and then collects the termination message.
// Timing is optional but if provided is [interval, timeout].
func WaitForJobTerminationMessage(ctx context.Context, t feature.T, name string, timing ...time.Duration) (string, error) {
	waitF := func(job *batchv1.Job) bool {
		return job.Status.Failed+job.Status.Succeeded > 0
	}
	if err := WaitForJobCondition(ctx, t, name, waitF, timing...); err != nil {
		return "", err
	}

	pod, err := GetJobPodByJobName(ctx, name)
	if err != nil {
		return "", err
	}
	return getFirstTerminationMessage(pod), nil
}

func IsCompleteJob(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsSucceededJob(job *batchv1.Job) bool {
	return IsCompleteJob(job) && !IsFailedJob(job)
}

func IsFailedJob(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func FailedMessage(job *batchv1.Job) string {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return fmt.Sprintf("[%s] %s", c.Reason, c.Message)
		}
	}
	return ""
}

// GetJobPod will find the Pod that belongs to the resource that created it.
// Uses label ""controller-uid  as the label selector. So, your job should
// tag the job with that label as the UID of the resource that's needing it.
// For example, if you create a storage object that requires us to create
// a notification for it, the controller should set the label on the
// Job responsible for creating the Notification for it with the label
// "controller-uid" set to the uid of the storage CR.
// TODO what is the use case for this function here?
// Deprecated, what is the use case here https://github.com/knative-sandbox/reconciler-test/issues/new?
func GetJobPod(ctx context.Context, kubeClientset kubernetes.Interface, namespace, uid, operation string) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)
	logger.Infof("Looking for Pod with UID: %q action: %q", uid, operation)
	matchLabels := map[string]string{
		"resource-uid": uid,
		"action":       operation,
	}
	labelSelector := &metav1.LabelSelector{
		MatchLabels: matchLabels,
	}
	pods, err := kubeClientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		logger.Infof("Found pod: %q", pod.Name)
		return &pod, nil
	}
	return nil, fmt.Errorf("Pod not found")
}

// GetJobPodByJobName will find the Pods that belong to that job. Each pod
// for a given job will have label called: "job-name" set to the job that
// it belongs to, so just filter by that.
// TODO what is the use case for this function here to require it to be public?
// TODO this function is tied to this job package and how it works, so make it private since
func GetJobPodByJobName(ctx context.Context, jobName string) (*corev1.Pod, error) {
	logger := logging.FromContext(ctx)
	namespace := environment.FromContext(ctx).Namespace()
	kube := kubeclient.Get(ctx)
	logger.Infof("Looking for Pod with jobname: %q", jobName)
	matchLabels := map[string]string{
		"job-name": jobName,
	}
	labelSelector := &metav1.LabelSelector{
		MatchLabels: matchLabels,
	}
	pods, err := kube.CoreV1().Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		logger.Infof("Found pod: %q", pod.Name)
		return &pod, nil
	}
	return nil, fmt.Errorf("Pod not found")
}

func getFirstTerminationMessage(pod *corev1.Pod) string {
	if pod != nil {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Terminated != nil && cs.State.Terminated.Message != "" {
				return cs.State.Terminated.Message
			}
		}
	}
	return ""
}

func registerImage(ctx context.Context, image string) error {
	reg := environment.RegisterPackage(image)
	_, err := reg(ctx, environment.FromContext(ctx))
	return err
}
