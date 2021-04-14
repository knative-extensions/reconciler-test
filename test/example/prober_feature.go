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

package example

import (
	"context"
	"fmt"

	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
)

func ProberFeature() *feature.Feature {
	f := &feature.Feature{Name: "Prober"}

	from := "yyyfrom"
	to := "yyyto"

	prober := eventshub.NewProber()

	// Configured the sender for how many events it will be sending.
	prober.SenderFullEvents(1)

	// Install the receiver, then the sender.
	f.Setup("install recorder", prober.ReceiverInstall(to))

	prober.AsKReference(to)
	_ = prober.SetTargetKRef(prober.AsKReference(to))

	f.Setup("install sender", prober.SenderInstall(from))
	f.Requirement("sender is done", prober.SenderDone(from))
	f.Requirement("receiver is done", prober.ReceiverDone(from, to))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the sender sent all events", prober.AssertSentAll(from)).
		Must("the recorder received all sent events", prober.AssertReceivedAll(from, to))

	return f
}

func ProberFeatureWithDrop() *feature.Feature {
	f := &feature.Feature{Name: "Prober with Drop"}

	from := "xxxfrom"
	to := "xxxto"

	prober := eventshub.NewProber()
	prober.ReceiversRejectFirstN(5)

	// Configured the sender for how many events it will be sending.
	prober.SenderFullEvents(6)

	// Install the receiver, then the sender.
	f.Setup("install recorder", prober.ReceiverInstall(to))

	prober.AsKReference(to)
	_ = prober.SetTargetKRef(prober.AsKReference(to))

	f.Setup("install sender", prober.SenderInstall(from))
	f.Requirement("sender is done", prober.SenderDone(from))
	f.Requirement("receiver is done", prober.ReceiverDone(from, to))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the sender sent all events", func(ctx context.Context, t feature.T) {
			events := prober.SentBy(ctx, from)
			if 6 != len(events) {
				t.Errorf("expected %q to have sent %d events, actually sent %d",
					from, 6, len(events))
			}
			for _, event := range events {
				switch event.Sent.SentId {
				case "1", "2", "3", "4", "5":
					if event.Response.StatusCode/100 != 4 {
						t.Errorf("For %s, expected 4xx response, got %d", event.Sent.SentId, event.Response.StatusCode)
					}
				case "6":
					if event.Response.StatusCode/100 != 2 {
						t.Errorf("For %s, expected 2xx response, got %d", event.Sent.SentId, event.Response.StatusCode)
					}
				}
			}
		}).
		Must("the recorder received all sent events", prober.AssertReceivedOrRejectedAll(from, to))

	return f
}

func ProberFeatureYAML() *feature.Feature {
	f := &feature.Feature{Name: "Prober with YAML"}

	from := "zzzfrom"
	to := "zzzto"

	prober := eventshub.NewProber()

	f.Setup("install recorder", prober.ReceiverInstall(to))

	_ = prober.SetTargetKRef(prober.AsKReference(to))

	// Locally, tell the test what to expect
	if err := prober.ExpectYAMLEvents("https://raw.githubusercontent.com/cloudevents/conformance/v0.2.0/yaml/v1.0/v1_minimum.yaml"); err != nil {
		panic(fmt.Errorf("can not load event files: %s", err))
	}
	// Configured the Sender to send the same events.
	prober.SenderEventsFromURI("https://raw.githubusercontent.com/cloudevents/conformance/v0.2.0/yaml/v1.0/v1_minimum.yaml")

	f.Setup("install sender with yaml events", prober.SenderInstall(from))

	f.Requirement("sender is done", prober.SenderDone(from))
	f.Requirement("receiver is done", prober.ReceiverDone(from, to))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the sender sent all events", prober.AssertSentAll(from)).
		Must("the recorder received all sent events", prober.AssertReceivedAll(from, to))

	return f
}
