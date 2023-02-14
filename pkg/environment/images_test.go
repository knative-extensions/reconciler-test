/*
Copyright 2022 The Knative Authors

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

package environment

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"knative.dev/reconciler-test/pkg/images/file"
)

func TestProduceImages(t *testing.T) {

	ctx := context.Background()

	readImages := func(images map[string]string) string {
		return images["x"]
	}
	writeImages := func(images map[string]string) {
		images["x"] = "x"
	}

	ctx, err := WithImages(map[string]string{
		"y": "y",
	})(ctx, nil)
	require.Nil(t, err)

	var wg sync.WaitGroup
	for x := 0; x < 100; x++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			images, err := ProduceImages(ctx)
			require.Nil(t, err)
			writeImages(images)
			i := readImages(images)
			t.Log("Image", i)
		}()
	}

	wg.Wait()
}

func TestWithImageProducer(t *testing.T) {

	ctx := context.Background()

	ctx, err := WithImageProducer(file.ImageProducer("testdata/images.yaml"))(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	ip := GetImageProducer(ctx)

	tt := []struct {
		key   string
		value string
	}{
		{
			key:   "knative.dev/reconciler-test/cmd/eventshub",
			value: "quay.io/myregistry/eventshub",
		},
	}

	for _, tc := range tt {
		got, err := ip(ctx, tc.key)
		if err != nil {
			t.Fatal(err)
		}

		if got != tc.value {
			t.Errorf("expected value %s, got %s", tc.value, got)
		}
	}
}
