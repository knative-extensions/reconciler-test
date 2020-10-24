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

package producer

import (
	"context"
	"fmt"
	"knative.dev/pkg/network"
	"knative.dev/reconciler-test/rigging"
	riggingv2 "knative.dev/reconciler-test/rigging/v2"
	"knative.dev/reconciler-test/rigging/v2/example/config"
	"testing"
)

func init() {
	rigging.RegisterPackage(
		"knative.dev/reconciler-test/rigging/v2/example/cmd/producer",
	)
}

func Install(sendCount int, sinkName string) riggingv2.PreConFn {
	return func(ctx context.Context, t *testing.T) {
		env := riggingv2.EnvFromContext(ctx)

		if err := config.InstallLocalYaml(ctx, map[string]interface{}{
			"count": fmt.Sprint(sendCount),
			"sink": fmt.Sprintf("http://%s.%s.svc.%s",
				sinkName, env.Namespace(), network.GetClusterDomainName()),
		}); err != nil {
			t.Fatal(err)
		}
	}
}
