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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/reconciler-test/pkg/k8s"

	. "github.com/cloudevents/sdk-go/v2/test"

	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/feature"
)

func RecorderFeature() *feature.Feature {
	svc := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

	from := feature.MakeRandomK8sName("sender")
	to := feature.MakeRandomK8sName("recorder")

	f := new(feature.Feature)

	event := FullEvent()

	f.Setup("install recorder", eventshub.Install(to, eventshub.StartReceiver))
	f.Setup("install sender", eventshub.Install(from, eventshub.StartSender(to), eventshub.InputEvent(event)))

	f.Requirement("recorder is addressable", k8s.IsAddressable(svc, to, time.Second, 30*time.Second))

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			func(ctx context.Context, t *testing.T) {
				eventshub.StoreFromContext(ctx, to).AssertExact(1, eventshub.MatchEvent(HasId(event.ID())))
			})

	return f
}
