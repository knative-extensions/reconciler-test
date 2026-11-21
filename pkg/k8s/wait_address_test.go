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

package k8s_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
	"knative.dev/pkg/injection/clients/dynamicclient"

	"knative.dev/reconciler-test/pkg/k8s"
)

var brokersGVR = schema.GroupVersionResource{Group: "eventing.knative.dev", Version: "v1", Resource: "brokers"}

func broker(url string) *unstructured.Unstructured {
	b := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "eventing.knative.dev/v1",
		"kind":       "Broker",
		"metadata":   map[string]interface{}{"name": "b", "namespace": "ns"},
	}}
	if url != "" {
		b.Object["status"] = map[string]interface{}{
			"address": map[string]interface{}{"url": url},
		}
	}
	return b
}

// withGets returns a context whose dynamic client answers the n-th get of the
// broker with responses[n], repeating the last one, and a counter of gets.
func withGets(responses ...func() (runtime.Object, error)) (context.Context, func() int) {
	gets := 0
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{brokersGVR: "BrokerList"})
	client.PrependReactor("get", "brokers", func(clienttesting.Action) (bool, runtime.Object, error) {
		respond := responses[min(gets, len(responses)-1)]
		gets++
		obj, err := respond()
		return true, obj, err
	})
	return context.WithValue(context.Background(), dynamicclient.Key{}, client), func() int { return gets }
}

// A sender is often installed while its target is still being set up, so the
// wait has to get past a missing resource, transient API errors and a
// resource that has no address yet.
func TestWaitForNamespacedAddress(t *testing.T) {
	ctx, gets := withGets(
		func() (runtime.Object, error) { return nil, apierrors.NewNotFound(brokersGVR.GroupResource(), "b") },
		func() (runtime.Object, error) {
			return nil, apierrors.NewInternalError(errors.New("etcd leader changed"))
		},
		func() (runtime.Object, error) { return broker(""), nil },
		func() (runtime.Object, error) { return broker("http://broker.ns.svc"), nil },
	)

	addr, err := k8s.WaitForNamespacedAddress(ctx, brokersGVR, "b", "ns", time.Millisecond, 5*time.Second)
	if err != nil {
		t.Fatalf("WaitForNamespacedAddress: %v", err)
	}
	if got, want := addr.URL.String(), "http://broker.ns.svc"; got != want {
		t.Errorf("address = %s, want %s", got, want)
	}
	if gets() != 4 {
		t.Errorf("gets = %d, want 4", gets())
	}
}

func TestWaitForNamespacedAddressTimesOut(t *testing.T) {
	ctx, _ := withGets(func() (runtime.Object, error) { return broker(""), nil })

	_, err := k8s.WaitForNamespacedAddress(ctx, brokersGVR, "b", "ns", time.Millisecond, 50*time.Millisecond)
	if !wait.Interrupted(err) {
		t.Errorf("err = %v, want a timeout", err)
	}
}

func TestWaitForNamespacedAddressStopsOnPermanentError(t *testing.T) {
	ctx, gets := withGets(func() (runtime.Object, error) {
		return nil, apierrors.NewForbidden(brokersGVR.GroupResource(), "b", errors.New("denied"))
	})

	_, err := k8s.WaitForNamespacedAddress(ctx, brokersGVR, "b", "ns", time.Millisecond, 5*time.Second)
	if !apierrors.IsForbidden(err) {
		t.Errorf("err = %v, want Forbidden", err)
	}
	if gets() != 1 {
		t.Errorf("gets = %d, want 1", gets())
	}
}
