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
	"knative.dev/reconciler-test/resources/svc"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/state"

	// Dot import the eventshub asserts and sdk-go test packages to include all the assert utilities
	. "github.com/cloudevents/sdk-go/v2/test"
	. "knative.dev/reconciler-test/pkg/eventshub/assert"
)

func RecorderFeature() *feature.Feature {
	f := &feature.Feature{Name: "Record"}

	f.Setup("create an event", func(ctx context.Context, t feature.T) {
		state.SetOrFail(ctx, t, "event", FullEvent())
	})

	f.Setup("install recorder", func(ctx context.Context, t feature.T) {
		to := feature.MakeRandomK8sName("recorder")
		state.SetOrFail(ctx, t, "to", to)
		eventshub.Install(to, eventshub.StartReceiver)(ctx, t)
	})

	f.Setup("install sender", func(ctx context.Context, t feature.T) {
		to := state.GetStringOrFail(ctx, t, "to")
		var event cloudevents.Event
		state.GetOrFail(ctx, t, "event", &event)
		from := feature.MakeRandomK8sName("sender")
		eventshub.Install(from, eventshub.StartSender(to), eventshub.InputEvent(event))(ctx, t)
	})

	f.Requirement("recorder is addressable", func(ctx context.Context, t feature.T) {
		to := state.GetStringOrFail(ctx, t, "to")
		k8s.IsAddressable(svc.GVR(), to, time.Second, 30*time.Second)(ctx, t)
	})

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			func(ctx context.Context, t feature.T) {
				to := state.GetStringOrFail(ctx, t, "to")
				var event cloudevents.Event
				state.GetOrFail(ctx, t, "event", &event)
				OnStore(to).MatchEvent(HasId(event.ID())).Exact(1)(ctx, t)
			})

	return f
}

func RecorderFeatureYAML() *feature.Feature {
	f := &feature.Feature{Name: "Record"}

	f.Setup("install recorder", func(ctx context.Context, t feature.T) {
		to := feature.MakeRandomK8sName("recorder")
		state.SetOrFail(ctx, t, "to", to)
		eventshub.Install(to, eventshub.StartReceiver)(ctx, t)
	})

	f.Setup("install sender with yaml events", func(ctx context.Context, t feature.T) {
		to := state.GetStringOrFail(ctx, t, "to")
		from := feature.MakeRandomK8sName("sender")
		eventshub.Install(from, eventshub.StartSender(to),
			eventshub.InputYAML("https://raw.githubusercontent.com/cloudevents/conformance/v0.2.0/yaml/v1.0/v1_minimum.yaml"))(ctx, t)
	})

	f.Requirement("recorder is addressable", func(ctx context.Context, t feature.T) {
		to := state.GetStringOrFail(ctx, t, "to")
		k8s.IsAddressable(svc.GVR(), to, time.Second, 30*time.Second)
	})

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			func(ctx context.Context, t feature.T) {
				to := state.GetStringOrFail(ctx, t, "to")
				OnStore(to).MatchEvent(HasId("conformance-0001")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0002")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0003")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0004")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0005")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0006")).Exact(1)(ctx, t)
			})

	return f
}
