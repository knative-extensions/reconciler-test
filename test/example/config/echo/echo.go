/*
Copyright 2020 The Knative Authors

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

package echo

import (
	"context"
	"embed"

	"knative.dev/reconciler-test/pkg/environment"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/manifest"
)

//go:embed *.yaml
var templates embed.FS

// Output is the base output we can expect from a echo job.
type Output struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

func Images() feature.Option {
	return func(ctx context.Context) (context.Context, error) {
		images := manifest.ImagesFromFS(ctx, templates)
		opt := environment.RegisterPackage(images...)
		return opt(ctx, environment.FromContext(ctx))
	}
}

func Install(name, message string) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		if _, err := manifest.InstallYamlFS(ctx, templates, map[string]interface{}{
			"name":    name,
			"message": message,
		}); err != nil {
			t.Fatal(err)
		}
	}
}
