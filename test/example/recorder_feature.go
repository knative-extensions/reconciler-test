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

package example

import (
	"context"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"

	// Dot import the eventshub asserts and sdk-go test packages to include all the assert utilities
	. "github.com/cloudevents/sdk-go/v2/test"
	. "knative.dev/reconciler-test/pkg/eventshub/assert"
)

func RecorderFeature() *feature.Feature {
	svc := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

	f := new(feature.Feature)

	f.Setup("create an event", func(ctx context.Context, t *testing.T) {
		t.Helper()
		f.Save(ctx, t, "event", FullEvent())
	})

	f.Setup("install recorder", func(ctx context.Context, t *testing.T) {
		t.Helper()
		to := feature.MakeRandomK8sName("recorder")
		eventshub.Install(to, eventshub.StartReceiver)
		f.Save(ctx, t, "to", to)
	})

	f.Setup("install sender", func(ctx context.Context, t *testing.T) {
		t.Helper()
		var to string
		var event cloudevents.Event
		f.Load(ctx, t, "to", &to)
		f.Load(ctx, t, "event", &event)
		from := feature.MakeRandomK8sName("sender")
		eventshub.Install(from, eventshub.StartSender(to), eventshub.InputEvent(event))
	})

	f.Requirement("recorder is addressable", func(ctx context.Context, t *testing.T) {
		t.Helper()
		var to string
		f.Load(ctx, t, "to", &to)
		k8s.IsAddressable(svc, to, time.Second, 30*time.Second)
	})

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			func(ctx context.Context, t *testing.T) {
				t.Helper()
				var to string
				var event cloudevents.Event
				f.Load(ctx, t, "to", &to)
				f.Load(ctx, t, "event", &event)
				OnStore(to).MatchEvent(HasId(event.ID())).Exact(1)
			})

	return f
}
