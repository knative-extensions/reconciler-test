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
	"time"

	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/resources/service"

	// Dot import the eventshub asserts and sdk-go test packages to include all the assert utilities
	. "github.com/cloudevents/sdk-go/v2/test"

	. "knative.dev/reconciler-test/pkg/eventshub/assert"
)

func RecorderFeature() *feature.Feature {
	f := &feature.Feature{Name: "Record"}

	to := feature.MakeRandomK8sName("recorder")
	from := feature.MakeRandomK8sName("sender")
	event := FullEvent()

	f.Setup("install recorder", eventshub.Install(to, eventshub.StartReceiver))
	f.Setup("recorder is addressable", k8s.IsAddressable(service.GVR(), to, time.Second, 30*time.Second))

	f.Requirement("install sender", eventshub.Install(from, eventshub.StartSender(to), eventshub.InputEvent(event)))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			OnStore(to).MatchEvent(HasId(event.ID())).Exact(1),
		)

	return f
}

func RecorderFeatureYAML() *feature.Feature {
	f := &feature.Feature{Name: "Record"}

	to := feature.MakeRandomK8sName("recorder")
	from := feature.MakeRandomK8sName("sender")

	f.Setup("install recorder", eventshub.Install(to, eventshub.StartReceiver))
	f.Setup("recorder is addressable", k8s.IsAddressable(service.GVR(), to, time.Second, 30*time.Second))

	f.Requirement("install sender with yaml events", eventshub.Install(from,
		eventshub.StartSender(to),
		eventshub.InputYAML("https://raw.githubusercontent.com/cloudevents/conformance/v0.2.0/yaml/v1.0/v1_minimum.yaml"),
	))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			func(ctx context.Context, t feature.T) {
				OnStore(to).MatchEvent(HasId("conformance-0001")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0002")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0003")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0004")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0005")).Exact(1)(ctx, t)
				OnStore(to).MatchEvent(HasId("conformance-0006")).Exact(1)(ctx, t)
			})

	return f
}
