package feature

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"knative.dev/reconciler-test/pkg/state"
)

func TestNewFeature(t *testing.T) {
	f := NewFeature()
	require.Equal(t, "TestNewFeature", f.Name)
}

func ExampleNewFeature() {
	f := NewFeature()
	f.State = &state.KVStore{}
	f.Reference(corev1.ObjectReference{
		Kind:       "Pod",
		Namespace:  "ns",
		Name:       "name",
		APIVersion: "v1",
	})

	f.Setup("step 1", func(ctx context.Context, t T) {})
	_ = f.State.Set(context.Background(), "key", "value")

	b, err := f.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	// Output:
	// {
	//  "name": "ExampleNewFeature",
	//  "steps": [
	//   {
	//    "name": "step 1",
	//    "states": "Any",
	//    "levels": "All",
	//    "timing": "Setup"
	//   }
	//  ],
	//  "state": {
	//   "key": "\"value\""
	//  },
	//  "refs": [
	//   {
	//    "kind": "Pod",
	//    "namespace": "ns",
	//    "name": "name",
	//    "apiVersion": "v1"
	//   }
	//  ]
	// }
}
