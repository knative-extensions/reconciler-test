/*
Copyright 2020 The Rigging Authors

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
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/meta"
	"knative.dev/reconciler-test/rigging"
	"knative.dev/reconciler-test/rigging/pkg/runner"
)

func init() {
	rigging.RegisterPackage("knative.dev/reconciler-test/rigging/cmd/testbed")
}

// BedTestImpl a very simple testbed test implementation.
func BedTestImpl(t *testing.T) {
	opts := []rigging.Option{}

	rig, err := rigging.NewInstall(opts, []string{"testbed"}, map[string]string{"echo": "hello world"})

	// Uninstall deferred.
	defer func() {
		time.Sleep(time.Minute * 2)
		if err := rig.Uninstall(); err != nil {
			t.Errorf("failed to uninstall, %s", err)
		}
	}()

	if err != nil {
		t.Fatalf("failed to create rig, %s", err)
	}
	t.Logf("Created a new testing rig at namespace %s.", rig.Namespace())

	refs := rig.Objects()
	for _, r := range refs {
		k := r.GroupVersionKind()
		gvk, _ := meta.UnsafeGuessKindToResource(k)
		t.Log(k, "--> UnsafeGuessKindToResource:", gvk)

		msg, err := rig.WaitForReadyOrDone(r, 45*time.Second)
		if err != nil {
			t.Fatalf("failed to wait for ready or done, %s", err)
		}
		if msg == "" {
			t.Error("No terminating message from the pod")
			return
		} else {
			out := &runner.Output{}

			t.Log("Got message from resource:", msg)

			org, _ := rig.ResourceOriginal(r)
			now, _ := rig.ResourceNow(r)

			if org != nil && now != nil {
				if diff := cmp.Diff(org, now); diff != "" {
					t.Log("FYI, diff on", r, diff)
				} else {
					t.Log("org or now are the same.")
				}
			} else {
				t.Log("org or now are nil")
				t.Logf("org: %v", org)
				t.Logf("now: %v", now)
			}

			if err := json.Unmarshal([]byte(msg), out); err != nil {
				t.Error(err)
				return
			}
			if !out.Success {
				if logs, err := rig.LogsFor(r); err != nil {
					t.Error(err)
				} else {
					t.Fatalf("failed with: %s\n", logs)
				}
				return
			}
		}
	}
}
