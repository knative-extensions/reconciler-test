/*
 * Copyright 2023 The Knative Authors
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

package service

import (
	corev1 "k8s.io/api/core/v1"
	"knative.dev/reconciler-test/pkg/manifest"
)

func WithAnnotations(annotations map[string]interface{}) manifest.CfgFn {
	return func(cfg map[string]interface{}) {
		if annotations != nil {
			cfg["annotations"] = annotations
		}
	}
}

func WithLabels(labels map[string]string) manifest.CfgFn {
	return func(cfg map[string]interface{}) {
		if labels != nil {
			cfg["labels"] = labels
		}
	}
}

func WithSelectors(selectors map[string]string) manifest.CfgFn {
	return func(cfg map[string]interface{}) {
		if selectors != nil {
			cfg["selectors"] = selectors
		}
	}
}

func WithType(serviceType corev1.ServiceType) manifest.CfgFn {
	return func(cfg map[string]interface{}) {
		cfg["type"] = serviceType
	}
}

func WithExternalName(externalName string) manifest.CfgFn {
	return func(cfg map[string]interface{}) {
		cfg["externalName"] = externalName
	}
}
