/*
 * Copyright 2020 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"testing"

	"knative.dev/pkg/apis"
)

type UpstreamStruct struct {
	Field1  string
	Nested1 UpstreamStruct1
}

type UpstreamStruct1 struct {
	Field1_1 string
	UpstreamStruct2
}

type UpstreamStruct2 struct {
	Field2_1 string
}

type downsstreamStruct struct {
	UpstreamStruct
	Field2 string
}

func (*UpstreamStruct) Validate() *apis.FieldError {
	return nil
}

func (*UpstreamStruct1) Validate() *apis.FieldError {
	return nil
}

func (*UpstreamStruct2) Validate() *apis.FieldError {
	return nil
}

func (*downsstreamStruct) Validate() *apis.FieldError {
	return nil
}

func TestGetConfig(t *testing.T) {
	v1 := downsstreamStruct{
		UpstreamStruct: UpstreamStruct{},
		Field2:         "value2",
	}

	v2 := downsstreamStruct{
		UpstreamStruct: UpstreamStruct{},
		Field2:         "value2",
	}

	v3 := downsstreamStruct{
		UpstreamStruct: UpstreamStruct{
			Field1: "value1",
		},
		Field2: "value2",
	}

	tests := []struct {
		name     string
		value    Config
		path     string
		expected Config
	}{
		{
			name:     "empty path",
			value:    &v1,
			path:     "",
			expected: &v1,
		},
		{
			name:     "path to nothing",
			value:    &v2,
			path:     "path/to/nothing",
			expected: nil,
		},
		{
			name:     "nested one level, return anonymous config",
			value:    &v3,
			path:     "UpstreamStruct",
			expected: &v3.UpstreamStruct,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := GetConfig(tc.value, tc.path)

			if actual != tc.expected {
				t.Fatalf("expected %+v, got %+v", tc.expected, actual)
			}
		})

	}
}
