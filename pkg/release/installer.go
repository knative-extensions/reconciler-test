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
package release

import (
	"fmt"

	"knative.dev/reconciler-test/pkg/framework"
	"knative.dev/reconciler-test/pkg/manifest"
)

func (r Release) Install(rc framework.ResourceContext, version string) {
	baseArtifactURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/", r.Owner, r.Repository, version)

	for _, artifact := range r.Artifacts {
		rc.Apply(manifest.FromURL(fmt.Sprintf("%s/%s", baseArtifactURL, artifact)))
	}
}
