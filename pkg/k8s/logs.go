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
	"fmt"

	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func LogsFor(client kubernetes.Interface, namespace, name string) (string, error) {
	logs, err := client.CoreV1().
		Pods(namespace).
		GetLogs(name, &corev1.PodLogOptions{}).
		DoRaw(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get logs for %s/%s: %w", namespace, name, err)
	}

	return string(logs), nil
}

func LogsForNamespace(client kubernetes.Interface, namespace string) (map[string]string, error) {
	pods, err := client.CoreV1().
		Pods(namespace).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	logs := make(map[string]string, len(pods.Items))
	globalErr := multierr.Combine()

	for _, p := range pods.Items {
		l, err := LogsFor(client, namespace, p.Name)
		if err != nil {
			globalErr = multierr.Append(globalErr, err)
		}
		// Append pod name and logs
		logs[p.Name] = l
	}

	return logs, globalErr
}
