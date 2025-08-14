/*
Copyright 2025 The Knative Authors

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

package observability

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
)

const (
	otelCollectorName   = "reconciler-test"
	observabilityCmName = "config-observability"
)

var (
	otelCollectorMux      = sync.Mutex{}
	otelCollectorResource = schema.GroupVersionResource{Group: "opentelemetry.io", Version: "v1alpha1", Resource: "opentelemetrycollector"}
	otelCollectorConfig   = `
receivers:
  otlp:
    protocols;
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
processors:

exporters:
  file/metrics:
    path: /data/metrics.json
  file/traces:
    path: /data/traces.json
service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: []
      exporters: [file/metrics]
    traces:
      receivers: [otlp]
      processors: []
      exporters: [file/traces]
`
)

func SetupObservability(ctx context.Context, observabilityNamespace string, configObservabilityNamespaces ...string) func() {
	return func() {
		dynamicClient := dynamicclient.Get(ctx)
		kubeClient := kubeclient.Get(ctx)

		logger := logging.FromContext(ctx)

		otelCollectorClient := dynamicClient.Resource(otelCollectorResource)
		_, err := otelCollectorClient.Namespace(observabilityNamespace).Get(ctx, otelCollectorName, metav1.GetOptions{})
		if err != nil && !apierrs.IsNotFound(err) {
			logger.Warnw("Failed to get OtelCollector, metrics and traces may not be collected", zap.Error(err))
			return
		}

		if err != nil && apierrs.IsNotFound(err) {
			otelCollectorMux.Lock()
			defer otelCollectorMux.Unlock()

			_, err := otelCollectorClient.Namespace(observabilityNamespace).Get(ctx, otelCollectorName, metav1.GetOptions{})
			if err != nil && !apierrs.IsNotFound(err) {
				logger.Warnw("Failed to get OtelCollector, metrics and traces may not be collected", zap.Error(err))
				return
			} else if err != nil {
				collector := &unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "opentelemetry.io/v1alpha1",
						"kind":       "OpenTelemetryCollector",
						"metadata": map[string]string{
							"name":      otelCollectorName,
							"namespace": observabilityNamespace,
						},
						"spec": map[string]any{
							"config": otelCollectorConfig,
							"volume": map[string]any{
								"name":     "file",
								"emptyDir": map[string]any{},
							},
							"volumeMounts": map[string]any{
								"name":      "file",
								"mountPath": "/data",
							},
						},
					},
				}

				_, err = otelCollectorClient.Create(ctx, collector, metav1.CreateOptions{})
				if err != nil && !apierrs.IsAlreadyExists(err) {
					logger.Warnw("Failed to create OtelCollector, metrics and traces will not be collected", zap.Error(err))
					return
				}
			}
		}

		for _, ns := range configObservabilityNamespaces {
			err = configureObservabilityCM(ctx, kubeClient, observabilityNamespace, ns)
		}
	}
}

func CleanupObservability(ctx context.Context, observabilityNamespace string) func() {
	return func() {
		kubeClient := kubeclient.Get(ctx)
		logger := logging.FromContext(ctx)

		labelSet := labels.Set{
			"app.kubernetes.io/instance": fmt.Sprintf("%s.%s", observabilityNamespace, otelCollectorName),
			"app.kubernetes.io/part-of":  "opentelemetry",
		}
		listOptions := metav1.ListOptions{
			LabelSelector: labelSet.AsSelector().String(),
		}

		pods, err := kubeClient.CoreV1().Pods(observabilityNamespace).List(ctx, listOptions)
		if err != nil {
			logger.Errorw("Failed to list opentelemetry collector pods, unable to save observability data", zap.Error(err))
			return
		}

		if len(pods.Items) < 1 {
			logger.Errorw("List of opentelemetry collector pods failed to find any pods", zap.String("LabelSelector", labelSet.AsSelector().String()))
		}

		artifacts := os.Getenv("ARTIFACTS")

		if _, err := os.Stat(fmt.Sprintf("%s/observability", artifacts)); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(fmt.Sprintf("%s/observability", artifacts), os.ModePerm)
			if err != nil {
				logger.Errorw("Failed to create directory for observability artifacts to be stored to", zap.Error(err))
				return
			}
		}

		metricsFile, err := os.OpenFile(fmt.Sprintf("%s/observability/metrics.json", artifacts), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Errorw("Failed to create file to write metrics to", zap.Error(err))
			return
		}
		defer metricsFile.Close()

		_, err = downloadFileFromPod(pods.Items[0], injection.GetConfig(ctx), kubeclient.Get(ctx), "/data/metrics.json", metricsFile)
		if err != nil {
			logger.Errorw("Failed to download metrics from otel collector", zap.Error(err))
			return
		}

		tracesFile, err := os.OpenFile(fmt.Sprintf("%s/observability/traces.json", artifacts), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Errorw("Failed to create file to write metrics to", zap.Error(err))
			return
		}
		defer tracesFile.Close()

		_, err = downloadFileFromPod(pods.Items[0], injection.GetConfig(ctx), kubeclient.Get(ctx), "/data/traces.json", tracesFile)
		if err != nil {
			logger.Errorw("Failed to download metrics from otel collector", zap.Error(err))
			return
		}
	}
}

func downloadFileFromPod(pod corev1.Pod, config *rest.Config, kubeClient kubernetes.Interface, filePath string, dest io.Writer) (int64, error) {
	reader, writer := io.Pipe()
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericiooptions.IOStreams{
				In:     nil,
				Out:    writer,
				ErrOut: bytes.NewBuffer([]byte{}),
			},
			Namespace: pod.Namespace,
			PodName:   pod.Name,
		},
		Executor:  &exec.DefaultRemoteExecutor{},
		Config:    config,
		PodClient: kubeClient.CoreV1(),
		Command:   []string{"/bin/cp", filePath, "/dev/stdout"},
	}

	go func(opt *exec.ExecOptions, writer *io.PipeWriter) {
		defer writer.Close()
		_ = opt.Run()

	}(options, writer)

	return io.Copy(dest, reader)
}

// configureObservabilityCM sets the correct options for the knative config-observability
// configmaps to connect to the otel collector deployed here.
//
// It assumes that the configmaps match the format in knative.dev/pkg/observability
// and are named config-observability
func configureObservabilityCM(ctx context.Context, kubeClient kubernetes.Interface, observabilityNamespace, configNamespace string) error {
	cm, err := kubeClient.CoreV1().ConfigMaps(configNamespace).Get(ctx, observabilityCmName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cm.Data["metrics-protocol"] = "grpc"
	cm.Data["metrics-endpoints"] = otelCollectorUrl(observabilityNamespace)
	cm.Data["runtime-profiling"] = "enabled"
	cm.Data["tracing-protocol"] = "grpc"
	cm.Data["tracing-endpoint"] = otelCollectorUrl(observabilityNamespace)
	cm.Data["tracing-sampling-rate"] = "1"

	_, err = kubeClient.CoreV1().ConfigMaps(configNamespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func otelCollectorUrl(namespace string) string {
	return fmt.Sprintf("http://%s-collector.%s:4317", otelCollectorName, namespace)
}
