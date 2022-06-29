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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
	testlog "knative.dev/reconciler-test/pkg/logging"

	"knative.dev/reconciler-test/pkg/feature"
)

type logEmitter struct {
	ctx       context.Context
	namespace string
}

// NewLogEmitter creates an Emitter that logs milestone events.
func NewLogEmitter(ctx context.Context, namespace string) Emitter {
	return &logEmitter{ctx: ctx, namespace: namespace}
}

func (l logEmitter) Environment(env map[string]string) {
	bytes, err := json.MarshalIndent(env, " ", " ")
	if err != nil {
		l.log().Fatal(err)
		return
	}
	l.log().Debug("Environment ", string(bytes))
}

func (l logEmitter) NamespaceCreated(namespace string) {
	l.log().Debug("Namespace created ", namespace)
}

func (l logEmitter) NamespaceDeleted(namespace string) {
	l.log().Debug("Namespace deleted ", namespace)
}

func (l logEmitter) TestStarted(feature string, t feature.T) {
	l.ctx = testlog.WithTestLogger(l.ctx, t)
	l.log().Debug(feature, " Test started")
}

func (l logEmitter) TestFinished(feature string, t feature.T) {
	l.log().Debug(feature, " Test Finished")
}

func (l logEmitter) StepsPlanned(feature string, steps map[feature.Timing][]feature.Step, t feature.T) {
	bytes, err := json.MarshalIndent(steps, " ", " ")
	if err != nil {
		l.log().Fatal(err)
		return
	}
	l.log().Debug(feature, " Steps Planned ", string(bytes))
}

func (l logEmitter) StepStarted(feature string, step *feature.Step, t feature.T) {
	bytes, err := json.MarshalIndent(step, " ", " ")
	if err != nil {
		l.log().Fatal(err)
		return
	}
	l.log().Debug(feature, " Step Started ", string(bytes))
}

func (l logEmitter) StepFinished(feature string, step *feature.Step, t feature.T) {
	bytes, err := json.MarshalIndent(step, " ", " ")
	if err != nil {
		l.log().Fatal(err)
		return
	}
	l.log().Debug(feature, " Step Finished ", string(bytes))
}

func (l logEmitter) TestSetStarted(featureSet string, t feature.T) {
	l.log().Debug(featureSet, " FeatureSet Started")
}

func (l logEmitter) TestSetFinished(featureSet string, t feature.T) {
	l.log().Debug(featureSet, " FeatureSet Finished")
}

func (l logEmitter) Finished() {
	l.log().Debug("Finished")
	l.dumpEvents()
}

func (l logEmitter) Exception(reason, messageFormat string, messageA ...interface{}) {
	l.log().Error("Exception ", reason, " ", fmt.Sprintf(messageFormat, messageA...))
}

func (l logEmitter) dumpEvents() {
	events, err := kubeclient.Get(l.ctx).CoreV1().Events(l.namespace).List(l.ctx, metav1.ListOptions{})
	if err != nil {
		l.log().Warn("failed to list events ", err)
		return
	}
	if len(events.Items) == 0 {
		l.log().Info("No events found")
		return
	}
	dump := l.newDumpFile()
	defer func() {
		_ = dump.Close()
	}()
	content, err := json.MarshalIndent(sortEventsByTime(events.Items), "", "  ")
	if err != nil {
		l.log().Fatal(err)
	}
	_, err = dump.Write(content)
	if err != nil {
		l.log().Fatal(err)
	}
	l.log().Infof("Events (%d) dump written to: %s",
		len(events.Items), dump.Name())
}

func (l logEmitter) newDumpFile() *os.File {
	artifacts := os.Getenv("ARTIFACTS")
	f, err := ioutil.TempFile(artifacts, "events-dump.*.json")
	if err != nil {
		l.log().Fatal(err)
	}
	return f
}

func sortEventsByTime(items []corev1.Event) []corev1.Event {
	sort.SliceStable(items, func(i, j int) bool {
		return eventTime(items[i]).Before(eventTime(items[j]))
	})
	return items
}

func eventTime(e corev1.Event) time.Time {
	// Some events might not contain last timestamp, in that case
	// we fall back to the event time.
	if e.LastTimestamp.Time.IsZero() {
		return e.EventTime.Time
	}
	return e.LastTimestamp.Time
}

func (l logEmitter) log() *zap.SugaredLogger {
	return logging.FromContext(l.ctx).With("namespace", l.namespace)
}
