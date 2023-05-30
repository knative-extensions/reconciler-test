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

package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/test/helpers"
)

func ExportLogs(client kubernetes.Interface, namespace, name string) error {
	logs, err := LogsFor(client, namespace, name)
	if err != nil {
		return err
	}
	return logs.Export()
}

func LogsFor(client kubernetes.Interface, namespace, name string) (Logs, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty")
	}
	// Get all pods in this namespace.
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	logs := make([]Log, 0, len(pods.Items))

	// Look for a pod with the name that was passed in inside the pod name.
	for _, pod := range pods.Items {
		if name == "" || strings.Contains(pod.Name, name) {
			// Collect all the logs from all the containers for this pod.
			for _, c := range pod.Spec.Containers {
				log := Log{
					PodName:       pod.Name,
					ContainerName: c.Name,
					PodNamespace:  pod.Namespace,
				}

				l, err := client.CoreV1().Pods(namespace).
					GetLogs(pod.Name, &corev1.PodLogOptions{Container: c.Name}).
					DoRaw(context.Background())
				if err != nil {
					log.Lines = fmt.Sprintf("Failed to get logs for %s/%s - %s: %v", pod.Namespace, pod.Name, c.Name, err)
				} else {
					log.Lines = string(l)
				}
				logs = append(logs, log)
			}
		}
	}

	return logs, nil
}

type Logs []Log

func (logs Logs) Export() error {
	var res error = nil
	for _, l := range logs {
		if err := l.export(); err != nil {
			// Do not fail fast, keep exporting logs and return the last non-nil error.
			res = err
		}
	}
	return res
}

type Log struct {
	PodNamespace  string
	PodName       string
	ContainerName string
	Lines         string
}

func (l Log) export() error {
	logs := filepath.Join(getLocalArtifactsDir(), "logs", l.PodNamespace)
	if err := helpers.CreateDir(logs); err != nil {
		return fmt.Errorf("error creating directory %q: %w", logs, err)
	}

	filename := fmt.Sprintf("%s_%s.log", l.PodName, l.ContainerName)
	fp := filepath.Join(logs, filename)
	f, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("error creating file %q: %w", fp, err)
	}
	defer f.Close()

	if _, err = f.Write([]byte(l.Lines)); err != nil {
		return fmt.Errorf("error writing logs into file %q: %w", fp, err)
	}
	return nil
}

// getLocalArtifactsDir gets the artifacts directory where prow looks for artifacts.
// By default, it will look at the env var ARTIFACTS.
func getLocalArtifactsDir() string {
	dir := os.Getenv("ARTIFACTS")
	if dir == "" {
		dir = "artifacts"
		log.Printf("Env variable ARTIFACTS not set. Using %s instead.", dir)
	}
	return dir
}
