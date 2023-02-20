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
	"knative.dev/reconciler-test/pkg/resources/job"
)

var (
	// WaitUntilJobDone
	// Deprecated
	WaitUntilJobDone = job.WaitUntilJobDone
	// WaitUntilJobSucceeded
	// Deprecated, use corresponding function in job package
	WaitUntilJobSucceeded = job.WaitUntilJobSucceeded
	// WaitUntilJobFailed
	// Deprecated, use corresponding function in job package
	WaitUntilJobFailed = job.WaitUntilJobFailed
	// WaitForJobCondition
	// Deprecated, use corresponding function in job package
	WaitForJobCondition = job.WaitForJobCondition
	// WaitForJobTerminationMessage
	// Deprecated, use corresponding function in job package
	WaitForJobTerminationMessage = job.WaitForJobTerminationMessage
	// IsJobComplete
	// Deprecated, use corresponding function in job package
	IsJobComplete = job.IsCompleteJob
	// IsJobSucceeded
	// Deprecated, use corresponding function in job package
	IsJobSucceeded = job.IsSucceededJob
	// IsJobFailed
	// Deprecated, use corresponding function in job package
	IsJobFailed = job.IsFailedJob
	// JobFailedMessage
	// Deprecated, use corresponding function in job package
	JobFailedMessage = job.FailedMessage
	// GetJobPod
	// Deprecated
	GetJobPod = job.GetJobPod
	// GetJobPodByJobName
	// Deprecated, use corresponding function in job package
	GetJobPodByJobName = job.GetJobPodByJobName
)
