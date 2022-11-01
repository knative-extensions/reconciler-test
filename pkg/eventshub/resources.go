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

package eventshub

import (
	"context"
	"embed"
	"strings"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
	pkgsecurity "knative.dev/pkg/test/security"
	"knative.dev/reconciler-test/pkg/environment"
	eventshubrbac "knative.dev/reconciler-test/pkg/eventshub/rbac"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/k8s"
	"knative.dev/reconciler-test/pkg/knative"
	"knative.dev/reconciler-test/pkg/manifest"
)

//go:embed *.yaml
var templates embed.FS

// Install starts a new eventshub with the provided name
// Note: this function expects that the Environment is configured with the
// following options, otherwise it will panic:
//
//	ctx, env := global.Environment(
//	  knative.WithKnativeNamespace("knative-namespace"),
//	  knative.WithLoggingConfig,
//	  knative.WithTracingConfig,
//	  k8s.WithEventListener,
//	)
func Install(name string, options ...EventsHubOption) feature.StepFn {
	return func(ctx context.Context, t feature.T) {
		if err := registerImage(ctx); err != nil {
			t.Fatalf("Failed to install eventshub image: %v", err)
		}
		env := environment.FromContext(ctx)
		namespace := env.Namespace()
		log := logging.FromContext(ctx)

		// Compute the user provided envs
		envs := make(map[string]string)
		if err := compose(options...)(ctx, envs); err != nil {
			log.Fatalf("Error while computing environment variables for eventshub: %s", err)
		}

		// eventshub needs tracing and logging config
		envs[ConfigLoggingEnv] = knative.LoggingConfigFromContext(ctx)
		envs[ConfigTracingEnv] = knative.TracingConfigFromContext(ctx)

		// Register the event info store to assert later the events published by the eventshub
		eventListener := k8s.EventListenerFromContext(ctx)
		registerEventsHubStore(ctx, eventListener, name, namespace)

		// Install ServiceAccount, Role, RoleBinding
		eventshubrbac.Install()(ctx, t)

		isReceiver := strings.Contains(envs["EVENT_GENERATORS"], "receiver")

		cfg := map[string]interface{}{
			"name":          name,
			"envs":          envs,
			"image":         ImageFromContext(ctx),
			"withReadiness": isReceiver,
		}

		restrictedMode, err := pkgsecurity.IsRestrictedPodSecurityEnforced(ctx, kubeclient.Get(ctx), namespace)
		if err != nil {
			log.Fatalf("Error while checking restricted pod security mode for namespace %s", namespace)
		}
		if restrictedMode {
			WithDefaultSecurityContext(cfg)
		}

		// Deploy
		if _, err := manifest.InstallYamlFS(ctx, templates, cfg); err != nil {
			log.Fatal(err)
		}

		podref, err := k8s.PodReference(namespace, name)
		if err != nil {
			log.Fatal(err)
		}
		k8s.WaitForPodRunningOrFail(ctx, t, name)
		k8s.WaitForReadyOrDoneOrFail(ctx, t, podref)

		// If the eventhubs starts an event receiver, we need to wait for the service endpoint to be synced
		if isReceiver {
			k8s.WaitForServiceEndpointsOrFail(ctx, t, name, 1)
			k8s.WaitForServiceReadyOrFail(ctx, t, name, "/health/ready")
		}
	}
}

func WithDefaultSecurityContext(cfg map[string]interface{}) {
	if _, set := cfg["podSecurityContext"]; !set {
		cfg["podSecurityContext"] = map[string]interface{}{}
	}
	podSecurityContext := cfg["podSecurityContext"].(map[string]interface{})
	podSecurityContext["runAsNonRoot"] = pkgsecurity.DefaultPodSecurityContext.RunAsNonRoot
	podSecurityContext["seccompProfile"] = map[string]interface{}{}
	seccompProfile := podSecurityContext["seccompProfile"].(map[string]interface{})
	seccompProfile["type"] = pkgsecurity.DefaultPodSecurityContext.SeccompProfile.Type

	if _, set := cfg["containerSecurityContext"]; !set {
		cfg["containerSecurityContext"] = map[string]interface{}{}
	}
	containerSecurityContext := cfg["containerSecurityContext"].(map[string]interface{})
	containerSecurityContext["allowPrivilegeEscalation"] =
		pkgsecurity.DefaultContainerSecurityContext.AllowPrivilegeEscalation
	containerSecurityContext["capabilities"] = map[string]interface{}{}
	capabilities := containerSecurityContext["capabilities"].(map[string]interface{})
	if len(pkgsecurity.DefaultContainerSecurityContext.Capabilities.Drop) != 0 {
		capabilities["drop"] = []string{}
		for _, drop := range pkgsecurity.DefaultContainerSecurityContext.Capabilities.Drop {
			capabilities["drop"] = append(capabilities["drop"].([]string), string(drop))
		}
	}
	if len(pkgsecurity.DefaultContainerSecurityContext.Capabilities.Add) != 0 {
		capabilities["add"] = []string{}
		for _, drop := range pkgsecurity.DefaultContainerSecurityContext.Capabilities.Drop {
			capabilities["add"] = append(capabilities["add"].([]string), string(drop))
		}
	}
}
