/*
Copyright 2026 The Knative Authors

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

package manifest

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

// The transient error classes retried here (apiserver 500s, unreachable
// admission webhooks, connection resets) are pinned by table tests in
// pkg/k8s (TestIsTransientError), using real errors observed in CI.

func TestApplyRetriesTransientError(t *testing.T) {
	fastRetry(t)

	client, creates := fakeClient(t, 2, apierrors.NewInternalError(errors.New("the server is currently unable to handle the request")))
	m := &YamlManifest{
		client:    client,
		log:       zap.NewNop().Sugar(),
		resources: []unstructured.Unstructured{testConfigMap("test-ns", "cm")},
	}

	if err := m.ApplyAll(); err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}
	if creates() != 3 {
		t.Errorf("expected create to be attempted 3 times (2 transient failures, then success), got %d", creates())
	}
}

func TestApplyDoesNotRetryPermanentError(t *testing.T) {
	fastRetry(t)

	invalidErr := apierrors.NewInvalid(
		schema.GroupKind{Group: "", Kind: "ConfigMap"}, "cm",
		field.ErrorList{field.Required(field.NewPath("data"), "")})
	client, creates := fakeClient(t, -1, invalidErr)
	m := &YamlManifest{
		client:    client,
		log:       zap.NewNop().Sugar(),
		resources: []unstructured.Unstructured{testConfigMap("test-ns", "cm")},
	}

	if err := m.ApplyAll(); err == nil {
		t.Fatal("expected ApplyAll to fail with the invalid-object error")
	}
	if creates() != 1 {
		t.Errorf("expected exactly 1 create attempt for a non-transient error, got %d", creates())
	}
}

// TestApplyRetryBudget pins the retry budget to at least ~30s so that short
// apiserver or admission-webhook blips (see TestIsTransientError) are absorbed
// instead of failing the test.
func TestApplyRetryBudget(t *testing.T) {
	total := time.Duration(0)
	d := applyRetry.Duration
	for i := 1; i < applyRetry.Steps; i++ {
		total += d
		d = time.Duration(float64(d) * applyRetry.Factor)
	}
	if total < 30*time.Second {
		t.Errorf("apply retry budget is only %v, want at least 30s to absorb transient apiserver/webhook blips", total)
	}
}

// fastRetry replaces the apply backoff with a near-instant one so retry tests
// do not sleep through the real budget.
func fastRetry(t *testing.T) {
	t.Helper()
	orig := applyRetry
	applyRetry = wait.Backoff{Steps: 5, Duration: time.Millisecond, Factor: 1}
	t.Cleanup(func() { applyRetry = orig })
}

// fakeClient returns a fake dynamic client whose "create" verb fails with
// failErr for the first failTimes calls (-1: always). The second return value
// reports how many create calls were made.
func fakeClient(t *testing.T, failTimes int, failErr error) (*dynamicfake.FakeDynamicClient, func() int) {
	t.Helper()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(), map[schema.GroupVersionResource]string{gvr: "ConfigMapList"})

	var creates int
	client.PrependReactor("create", "configmaps", func(action clienttesting.Action) (bool, runtime.Object, error) {
		creates++
		if failTimes < 0 || creates <= failTimes {
			return true, nil, failErr
		}
		return false, nil, nil // delegate to the object tracker, which succeeds
	})
	return client, func() int { return creates }
}

func testConfigMap(namespace, name string) unstructured.Unstructured {
	u := unstructured.Unstructured{}
	u.SetAPIVersion("v1")
	u.SetKind("ConfigMap")
	u.SetNamespace(namespace)
	u.SetName(name)
	return u
}
