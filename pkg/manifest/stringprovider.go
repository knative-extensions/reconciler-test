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

package manifest

import (
	"io/ioutil"
)

type stringProvider struct {
	str string
}

func (p *stringProvider) GetPath() (string, error) {
	// TODO: check string is valid YAML?

	f, err := ioutil.TempFile("", "manifest-*.yaml")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.WriteString(p.str)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

func (p *stringProvider) Recursive() bool {
	return false
}
