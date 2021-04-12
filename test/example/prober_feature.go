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
	"fmt"

	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
)

func ProberFeature() *feature.Feature {
	f := &feature.Feature{Name: "Prober"}

	prober := eventshub.NewProber()

	// Configured the sender for how many events it will be sending.
	prober.SenderFullEvents(1)

	// Install the receiver, then the sender.
	f.Setup("install recorder", prober.ReceiverInstall("to"))

	prober.AsKRef("to")
	_ = prober.SetTargetKRef(prober.AsKRef("to"))

	f.Setup("install sender", prober.SenderInstall("from"))
	f.Requirement("sender is done", prober.SenderDone("from"))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the sender sent all events", prober.AssertSentAll("from")).
		Must("the recorder received all sent events", prober.AssertReceivedAll("from", "to"))

	return f
}

func ProberFeatureYAML() *feature.Feature {
	f := &feature.Feature{Name: "Prober with YAML"}

	prober := eventshub.NewProber()

	f.Setup("install recorder", prober.SenderInstall("recorder"))

	_ = prober.SetTargetKRef(prober.AsKRef("recorder"))

	// Locally, tell the test what to expect
	if err := prober.ExpectYAMLEvents("https://raw.githubusercontent.com/cloudevents/conformance/v0.2.0/yaml/v1.0/v1_minimum.yaml"); err != nil {
		panic(fmt.Errorf("can not load event files: %s", err))
	}
	// Configured the Sender to send the same events.
	prober.SenderEventsFromURI("https://raw.githubusercontent.com/cloudevents/conformance/v0.2.0/yaml/v1.0/v1_minimum.yaml")

	f.Setup("install sender with yaml events", prober.SenderInstall("sender"))

	f.Requirement("sender is done", prober.SenderDone("sender"))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the sender sent all events", prober.AssertSentAll("sender")).
		Must("the recorder received all sent events", prober.AssertReceivedAll("sender", "recorder"))

	return f
}
