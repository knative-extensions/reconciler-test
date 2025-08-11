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

package milestone

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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
	"knative.dev/reconciler-test/pkg/feature"
)

const (
	otelCollectorName = "reconciler-test"
)

var (
	otelCollectorMux      sync.Mutex
	otelCollectorResource = schema.GroupVersionResource{Group: "opentelemetry.io", Version: "v1alpha1", Resource: "opentelemetrycollector"}
	otelCollectorConfig   = `
receivers:
  otlp:
    protocols:
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

// LogEmitter is an Emitter that logs milestone events.
type ObservabilityEmitter struct {
	ctx                    context.Context
	dynamicClient          dynamic.Interface
	kubeClient             kubernetes.Interface
	config                 *rest.Config
	observabilityNamespace string
	logger                 *zap.SugaredLogger
}

// NewLogEmitter creates an Emitter that logs milestone events.
func NewObservabilityGatherer(ctx context.Context, observabilityNamespace string, t feature.T) (*ObservabilityEmitter, error) {
	kubeClient := kubeclient.Get(ctx)
	dynamicClient := dynamicclient.Get(ctx)
	cfg := injection.GetConfig(ctx)

	// set up the otel collector
	otelCollectorClient := dynamicClient.Resource(otelCollectorResource)
	_, err := otelCollectorClient.Namespace(observabilityNamespace).Get(ctx, otelCollectorName, metav1.GetOptions{})
	if err != nil && !apierrs.IsNotFound(err) {
		return nil, err
	}

	if err != nil && apierrs.IsNotFound(err) {
		otelCollectorMux.Lock()
		defer otelCollectorMux.Unlock()

		// try again now that we have the lock
		_, err := otelCollectorClient.Namespace(observabilityNamespace).Get(ctx, otelCollectorName, metav1.GetOptions{})
		if err != nil && !apierrs.IsNotFound(err) {
			return nil, err
		} else if err != nil {
			// the collector definitely does not exist, let's creat it
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
				return nil, err
			}
		}
	}

	return &ObservabilityEmitter{
		ctx:                    ctx,
		kubeClient:             kubeClient,
		dynamicClient:          dynamicClient,
		config:                 cfg,
		observabilityNamespace: observabilityNamespace,
		logger:                 logging.FromContext(ctx),
	}, nil
}

func (o *ObservabilityEmitter) Environment(env map[string]string) {
}

func (o *ObservabilityEmitter) NamespaceCreated(namespace string) {
}

func (o *ObservabilityEmitter) NamespaceDeleted(namespace string) {
}

func (o *ObservabilityEmitter) TestStarted(feature string, t feature.T) {
}

func (o *ObservabilityEmitter) TestFinished(feature string, t feature.T) {
}

func (o *ObservabilityEmitter) StepsPlanned(feature string, steps map[feature.Timing][]feature.Step, t feature.T) {
}

func (o *ObservabilityEmitter) StepStarted(feature string, step *feature.Step, t feature.T) {
}

func (o *ObservabilityEmitter) StepFinished(feature string, step *feature.Step, t feature.T) {
}

func (o *ObservabilityEmitter) TestSetStarted(featureSet string, t feature.T) {
}

func (o *ObservabilityEmitter) TestSetFinished(featureSet string, t feature.T) {
}

func (o *ObservabilityEmitter) Finished(result Result) {
	// get collector pods
	// copy data files from collector pods to artefacts directory
	labelSet := labels.Set{
		"app.kubernetes.io/instance": fmt.Sprintf("%s.%s", o.observabilityNamespace, otelCollectorName),
		"app.kubernetes.io/part-of":  "opentelemetry",
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labelSet.AsSelector().String(),
	}
	pods, err := o.kubeClient.CoreV1().Pods(o.observabilityNamespace).List(o.ctx, listOptions)
	if err != nil {
		o.logger.Errorw("Failed to list opentelemetry collector pods, unable to save observability data", zap.Error(err))
		return
	}

	if len(pods.Items) < 1 {
		o.logger.Errorw("List of opentelemetry collector pods failed to find any pods", zap.String("LabelSelector", labelSet.AsSelector().String()))
		return
	}

	artifacts := os.Getenv("ARTIFACTS")

	if _, err := os.Stat(fmt.Sprintf("%s/observability", artifacts)); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(fmt.Sprintf("%s/observability", artifacts), os.ModePerm)
		if err != nil {
			o.logger.Errorw("Failed to create directory for observability artifacts to be stored to", zap.Error(err))
			return
		}
	}

	metricsFile, err := os.OpenFile(fmt.Sprintf("%s/observability/metrics.json", artifacts), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		o.logger.Errorw("Failed to create file to write metrics to", zap.Error(err))
		return
	}
	defer metricsFile.Close()

	_, err = o.downloadFileFromPod(&pods.Items[0], "/data/metrics.json", metricsFile)
	if err != nil {
		o.logger.Errorw("Failed to copy metrics from collector to artifacts dir", zap.Error(err))
		return
	}

	tracesFile, err := os.OpenFile(fmt.Sprintf("%s/observability/metrics.json", artifacts), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		o.logger.Errorw("Failed to create file to write traces to", zap.Error(err))
		return
	}
	defer tracesFile.Close()

	_, err = o.downloadFileFromPod(&pods.Items[0], "/data/traces.json", tracesFile)
	if err != nil {
		o.logger.Errorw("Failed to copy traces from collector to artifacts dir", zap.Error(err))
		return
	}
}

func (o *ObservabilityEmitter) Exception(reason, messageFormat string, messageA ...any) {
}

// adapted from https://github.com/nvanheuverzwijn/k8s-operator-examples/blob/master/backup-operator2/pkg/pod/podfile.go
func (o *ObservabilityEmitter) downloadFileFromPod(pod *corev1.Pod, path string, w io.Writer) (int64, error) {
	options := &exec.ExecOptions{}
	errOut := bytes.NewBuffer([]byte{})
	reader, writer := io.Pipe()

	options.StreamOptions = exec.StreamOptions{
		IOStreams: genericiooptions.IOStreams{
			In:     nil,
			Out:    writer,
			ErrOut: errOut,
		},
		Namespace: pod.Namespace,
		PodName:   pod.Name,
	}
	options.Executor = &exec.DefaultRemoteExecutor{}
	options.Config = o.config
	options.PodClient = o.kubeClient.CoreV1()
	options.Command = []string{"/bin/cp", path, "/dev/stdout"}

	go func(opt *exec.ExecOptions, writer *io.PipeWriter) {
		defer writer.Close()
		_ = opt.Run()
	}(options, writer)

	return io.Copy(w, reader)
}
