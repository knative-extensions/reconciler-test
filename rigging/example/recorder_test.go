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
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/network"
	"knative.dev/reconciler-test/pkg/observer"
	recorder_collector "knative.dev/reconciler-test/pkg/observer/recorder-collector"
	"knative.dev/reconciler-test/rigging"
)

func init() {
	rigging.RegisterPackage(
		"knative.dev/reconciler-test/images/recorder",
		"knative.dev/reconciler-test/rigging/example/cmd/producer",
	)
}

// RecorderTestImpl a very simple example test implementation.
//
func RecorderTestImpl(t *testing.T) {
	sendCount := 5
	opts := []rigging.Option{}

	rig, err := rigging.NewInstall(opts, []string{"producer", "recorder"}, map[string]string{
		"producerCount":     fmt.Sprint(sendCount),
		"producerSink":      "recorder",
		"clusterDomainName": network.GetClusterDomainName(),
	})
	if err != nil {
		t.Fatalf("failed to create rig, %s", err)
	}

	t.Logf("Created a new testing rig at namespace %s.", rig.Namespace())

	// Uninstall deferred.
	defer func() {
		if err := rig.Uninstall(); err != nil {
			t.Errorf("failed to uninstall, %s", err)
		}
	}()

	// TODO: need to validate set events.
	ctx := Context() // TODO: there needs to be a better way to do this.
	c := recorder_collector.New(ctx)

	from := duckv1.KReference{
		Kind:       "Namespace",
		Name:       "default",
		APIVersion: "v1",
	}

	obsName := "recorder-" + rig.Namespace()

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		events, err := c.List(ctx, from, func(ob observer.Observed) bool {
			return ob.Observer == obsName
		})
		if err != nil {
			return false, err
		}

		for i, e := range events {
			fmt.Printf("[%d]: seen by %q\n%s\n", i, e.Observer, e.Event)
		}

		got := len(events)
		want := sendCount
		if want != got {
			t.Logf("dod not observe the correct number of events, want: %d, got: %d", want, got)
			return false, nil
		} else {
			return true, nil
		}
	})
	if err != nil {
		t.Error("failed to observe the correct number of events, ", err)
	}

	// Pass!
}
