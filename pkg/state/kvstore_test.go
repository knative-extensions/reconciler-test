/*
Copyright 2021 The Knative Authors

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

package state

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestKVStore_Get(t *testing.T) {
	type test1 struct {
		Foo string
		Bar string
	}

	tests := []struct {
		name    string
		store   map[string]string
		key     string
		target  interface{}
		want    interface{}
		wantErr bool
	}{{
		name:  "nil store",
		store: nil,
		key:   "foo",
		target: func() interface{} {
			return ""
		}(),
		want:    "",
		wantErr: true,
	}, {
		name: "found",
		store: map[string]string{
			"foo": `"bar"`,
		},
		key: "foo",
		target: func() interface{} {
			return ""
		}(),
		want:    "bar",
		wantErr: false,
	}, {
		name: "found struct",
		store: map[string]string{
			"key": `{"Foo": "hello","Bar": "world"}`,
		},
		key: "key",
		target: func() interface{} {
			return test1{}
		}(),
		want: test1{
			Foo: "hello",
			Bar: "world",
		},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := KVStore{
				store: tt.store,
			}
			err := s.Get(context.Background(), tt.key, &tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if diff := cmp.Diff(tt.want, tt.target); diff != "" {
					t.Error("Unexpected difference on return value (-want, +got):", diff)
				}
			}
		})
	}
}
