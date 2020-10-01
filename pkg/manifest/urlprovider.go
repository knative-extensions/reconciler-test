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
	"net/http"
	"net/url"
	"path"
	"strings"
)

type urlProvider struct {
	url       string
	recursive bool
}

func (p *urlProvider) GetPath() (string, error) {
	if isURL(p.url) {
		// Fetches manifest(s)
		resp, err := http.Get(p.url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		u, _ := url.Parse(p.url) // no error (see isURL)
		stem := strings.TrimRight(path.Base(u.Path), ".yaml")

		f, err := ioutil.TempFile("", stem+"*.yaml")
		if err != nil {
			return "", err
		}
		defer f.Close()

		buffer, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		_, err = f.Write(buffer)
		if err != nil {
			return "", err
		}
		return f.Name(), nil
	}

	return p.url, nil
}

func (p *urlProvider) Recursive() bool {
	return p.recursive
}
