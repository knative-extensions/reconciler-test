/*
Copyright 2023 The Knative Authors

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

package file

import (
	"context"
	"testing"
)

func TestImageProducer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	producer := ImageProducer("testdata/images.yaml")

	tt := []struct {
		key     string
		value   string
		wantErr bool
	}{
		{
			key:   "knative.dev/reconciler-test/cmd/eventshub",
			value: "quay.io/myregistry/eventshub",
		},
		{
			key:   "knative.dev/reconciler-test/cmd/eventshub2",
			value: "quay.io/myregistry/eventshub2",
		},
		{
			key:     "knative.dev/reconciler-test/cmd/eventshub3",
			wantErr: true,
		},
	}

	for _, tc := range tt {
		got, err := producer(ctx, tc.key)
		if tc.wantErr != (err != nil) {
			t.Fatal("want error", tc.wantErr, "error", err)
		}

		if !tc.wantErr && got != tc.value {
			t.Errorf("expected value %s, got %s", tc.value, got)
		}
	}
}
