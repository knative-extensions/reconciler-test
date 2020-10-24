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

package config

import (
	"context"
	"log"
	"os"
	"path"
	"runtime"

	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/reconciler-test/rigging/pkg/installer"
	yaml "knative.dev/reconciler-test/rigging/pkg/manifest"
	riggingv2 "knative.dev/reconciler-test/rigging/v2"
)

func InstallLocalYaml(ctx context.Context, base map[string]interface{}) error {
	env := riggingv2.EnvFromContext(ctx)
	cfg := env.TemplateConfig(base)

	pwd, _ := os.Getwd()
	log.Println("PWD: ", pwd)
	_, filename, _, _ := runtime.Caller(1)
	log.Println("FILENAME: ", filename)

	yamls := installer.ParseTemplates(path.Dir(filename), cfg)

	dynamicClient := dynamicclient.Get(ctx)

	manifest, err := yaml.NewYamlManifest(yamls, false, dynamicClient)
	if err != nil {
		return err
	}

	// 4. Apply yaml.
	if err := manifest.ApplyAll(); err != nil {
		return err
	}

	// TODO: likely want to save manifest?

	// Temp
	refs := manifest.References()
	log.Println("Created:")
	for _, ref := range refs {
		log.Println(ref)
	}
	return nil
}
