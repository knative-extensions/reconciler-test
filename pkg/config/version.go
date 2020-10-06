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
	"fmt"

	"knative.dev/pkg/apis"

	"github.com/blang/semver"
)

type VersionSpec struct {
	Version string `desc:"the resolved version (only needed if require is a range)"`
	Require string `desc:"the required version. Either a pin version, a version range or main"`
	Path    string `desc:"the root location of the component source code"`

	// ActualVersion is the version of the deployed component
	ActualVersion string `flag:"-"`

	// Parsed Required
	vrange semver.Range `flag:"-"`
}

// Validate checks VersionSpec is correct
func (spec *VersionSpec) Validate() *apis.FieldError {
	if spec.Require == "" {
		return apis.ErrMissingField("require")
	}

	if spec.Require == "main" {
		if spec.Path == "" {
			return apis.ErrMissingField("path")
		}
		return nil
	}

	r, err := semver.ParseRange(spec.Require)
	if err != nil {
		return apis.ErrInvalidValue(spec.Require, "require")
	}
	spec.vrange = r

	_, err = semver.Parse(spec.Require)
	if err == nil {
		if spec.Version != "" && spec.Version != spec.Require {
			return apis.ErrGeneric(fmt.Sprintf("mismatch required and resolved version (%s != %s)", spec.Version, spec.Require), "require", "version")

		}
		spec.Version = StripBuild(spec.Require)
		return nil
	}

	if spec.Version == "" {
		// TODO: dynamic version resolution
		return apis.ErrMissingField("version")
	}

	return nil
}

type CompareType int

const (
	// CompareDevel indicates both the resolved version and the compared version are "devel"
	CompareDevel = CompareType(iota)

	// CompareInvalidVersion indicates the compared version is invalid
	CompareInvalidVersion

	// CompareInRange indicates the compared version is in range of the required version
	CompareInRange

	// CompareOutOfRange indicates the compared version is not in the range of the required version
	CompareOutOfRange

	// CompareEmptyVersion indicates the compared version is empty
	CompareEmptyVersion
)

// Compare compares version against the spec
func (spec *VersionSpec) Compare(version string) CompareType {
	spec.ActualVersion = version
	if version == "" {
		return CompareEmptyVersion
	}

	if version == "devel" {
		if spec.Version == "devel" {
			return CompareDevel
		}
		return CompareOutOfRange
	}

	v, err := semver.ParseTolerant(version)
	if err != nil {
		return CompareInvalidVersion
	}

	if spec.vrange(v) {
		return CompareInRange
	}

	return CompareOutOfRange

	return CompareEmptyVersion
}

// StripBuild returns version without build meta-data
func StripBuild(version string) string {
	v, err := semver.Parse(version)
	if err != nil {
		panic(err)
	}
	v.Build = nil
	return v.String()
}
