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
	"fmt"
	riggingv2 "knative.dev/reconciler-test/rigging/v2"
	"knative.dev/reconciler-test/rigging/v2/example/config/producer"
	"knative.dev/reconciler-test/rigging/v2/example/config/recorder"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/reconciler-test/pkg/observer"
	recorder_collector "knative.dev/reconciler-test/pkg/observer/recorder-collector"
)

func RecorderFeature() *riggingv2.Feature {
	count := 5

	f := new(riggingv2.Feature)

	f.Precondition("install recorder", recorder.Install())
	f.Precondition("install producer", producer.Install(count, "recorder")) // TODO: it would have been nice to know the name of the recorder programmatically

	whoFn := func(namespace string) (string, duckv1.KReference) {
		from := duckv1.KReference{
			Kind:       "Namespace",
			Name:       namespace,
			APIVersion: "v1",
		}
		to := "recorder-" + namespace
		return to, from
	}

	f.Alpha("direct sending between a producer and a recorder").
		Must("the recorder received all sent events within the time",
			AssertDelivery(whoFn, count, 1*time.Second, 20*time.Second))

	return f
}

type whoFn = func(namespace string) (string, duckv1.KReference)

func AssertDelivery(who whoFn, count int, interval, timeout time.Duration) riggingv2.AssertFn {
	return func(ctx context.Context, t *testing.T) {
		to, from := who(riggingv2.EnvFromContext(ctx).Namespace())

		c := recorder_collector.New(ctx)
		if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
			events, err := c.List(ctx, from, func(ob observer.Observed) bool {
				fmt.Printf("[filter]: %q - %s : %t\n", ob.Observer, ob.Event.Type(), ob.Observer == to)
				return ob.Observer == to
			})
			if err != nil {
				return false, err
			}

			for i, e := range events {
				fmt.Printf("[%d]: seen by %q\n%s\n", i, e.Observer, e.Event)
			}

			got := len(events)
			want := count
			if want != got {
				t.Logf("did not observe the correct number of events, want: %d, got: %d", want, got)
				return false, nil
			} else {
				return true, nil
			}
		}); err != nil {
			t.Error("failed to observe the correct number of events, ", err)
		}
	}

}
