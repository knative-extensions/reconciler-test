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

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name     string
		value    downsstreamStruct
		path     string
		expected interface{}
	}{
		{
			name: "empty path",
			value: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{},
				Field2:         "value2",
			},
			path: "",
			expected: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{},
				Field2:         "value2",
			},
		},
		{
			name: "path to nothing",
			value: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{},
				Field2:         "value2",
			},
			path:     "path/to/nothing",
			expected: nil,
		},
		{
			name: "top-level field",
			value: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{},
				Field2:         "value2",
			},
			path:     "field2",
			expected: "value2",
		},
		{
			name: "nested one level, anonymous",
			value: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{
					Field1: "value1",
				},
				Field2: "value2",
			},
			path:     "field1",
			expected: "value1",
		},
		{
			name: "nested one level, return anonymous config",
			value: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{
					Field1: "value1",
				},
				Field2: "value2",
			},
			path: "UpstreamStruct",
			expected: UpstreamStruct{
				Field1: "value1",
			},
		},
		{
			name: "nested two levels, not anonymous",
			value: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{
					Field1: "value1",
					Nested1: UpstreamStruct1{
						Field1_1: "value3",
					},
				},
				Field2: "value2",
			},
			path:     "nested1/field1_1",
			expected: "value3",
		},
		{
			name: "nested three levels, not anonymous, anonymous",
			value: downsstreamStruct{
				UpstreamStruct: UpstreamStruct{
					Field1: "value1",
					Nested1: UpstreamStruct1{
						Field1_1: "value3",
						UpstreamStruct2: UpstreamStruct2{
							Field2_1: "value4",
						},
					},
				},
				Field2: "value2",
			},
			path:     "nested1/field2_1",
			expected: "value4",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := GetConfig(tc.value, tc.path)

			if actual != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})

	}
}
