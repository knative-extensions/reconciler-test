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

package recorder

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/pkg/observer"
	recorder_collector "knative.dev/reconciler-test/pkg/observer/recorder-collector"
)

func init() {
	environment.RegisterPackage(
		"knative.dev/reconciler-test/images/recorder",
	)
}

func Install(recorderName string) feature.PreConFn {
	return func(ctx context.Context, t *testing.T) {
		if _, err := manifest.InstallLocalYaml(ctx, map[string]interface{}{
			"recorderName": recorderName,
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func AssertDelivery(to string, count int, interval, timeout time.Duration) feature.AssertFn {
	return func(ctx context.Context, t *testing.T) {
		env := environment.FromContext(ctx)

		from := duckv1.KReference{
			Kind:       "Namespace",
			Name:       env.Namespace(),
			APIVersion: "v1",
		}

		c := recorder_collector.New(ctx)
		if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
			events, err := c.List(ctx, from, func(ob observer.Observed) bool {
				return ob.Observer == to
			})
			if err != nil {
				return false, err
			}

			for i, e := range events {
				t.Logf("[%d]: seen by %q\n%s\n", i, e.Observer, e.Event)
			}

			got := len(events)
			want := count
			if want != got {
				t.Logf("did not yet observe the correct number of events, want: %d, got: %d", want, got)
				return false, nil
			} else {
				return true, nil
			}
		}); err != nil {
			t.Error("failed to observe the correct number of events, ", err)
		}
	}

}
