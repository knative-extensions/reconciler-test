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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/test/example/config/echo"
)

func EchoFeature() *feature.Feature {
	f := &feature.Feature{Name: "EchoFeature"}

	f.ConfigureEnvironment(echo.Images())

	msg := fmt.Sprintf("hello %s", uuid.New())
	name := "echo" + uuid.New().String()

	f.Setup("install echo", echo.Install(name, msg))

	f.Requirement("echo job is finished", func(ctx context.Context, t feature.T) {
		if err := k8s.WaitUntilJobDone(ctx, t, name); err != nil {
			t.Errorf("failed to wait for job to finish, %s", err)
		}
	})

	f.Alpha("pull logs off a pod").
		Must("the echo pod must echo our message",
			func(ctx context.Context, t feature.T) {
				// The usage of WaitForJobTerminationMessage here explicitly sets the poll timings.
				log, err := k8s.WaitForJobTerminationMessage(ctx, t, name, time.Second, 30*time.Second)
				if err != nil {
					t.Error("failed to get termination message from pod, ", err)
				}

				out := &echo.Output{}
				if err := json.Unmarshal([]byte(log), out); err != nil {
					t.Error("failed to unmarshal pod log: ", log, err)
					return
				}

				if !out.Success {
					t.Error("failed with: \n", log)
					return
				}

				if out.Message != msg {
					t.Errorf("echo message does not match, wanted: %s, got: %s", msg, out.Message)
					return
				}
				t.Log("got our message echo'ed: ", out.Message)
			}).
		May("An example of a MAY", func(ctx context.Context, t feature.T) {
			t.Log("ran inside of a MAY")
		}).
		Should("An example of a SHOULD", func(ctx context.Context, t feature.T) {
			t.Log("ran inside of a SHOULD")
		})

	f.Teardown("cleanup", f.DeleteResources)

	return f
}

// EchoFeatureSet makes a feature set out of a few EchoFeatures for testing.
func EchoFeatureSet() *feature.FeatureSet {
	fs := &feature.FeatureSet{
		Name: "Echo Feature Wrapper (3x)",
		Features: []*feature.Feature{
			EchoFeature(),
			EchoFeature(),
			EchoFeature(),
		},
	}
	return fs
}
