#!/usr/bin/env bash

# Copyright 2023 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -Eeo pipefail

root_dir="$(dirname "$(dirname "$(readlink -f "${BASH_SOURCE[0]:-$0}")")")"
readonly root_dir

export CERT_MANAGER_NAMESPACE="cert-manager"

source "${root_dir}/vendor/knative.dev/hack/e2e-tests.sh"

function test_setup() {
  kubectl apply -f third_party/cert-manager/00-namespace.yaml

  timeout 600 bash -c 'until kubectl apply -f third_party/cert-manager/01-cert-manager.yaml; do sleep 5; done'
  wait_until_pods_running "$CERT_MANAGER_NAMESPACE" || fail_test "Failed to install cert manager"

  timeout 600 bash -c 'until kubectl apply -f third_party/cert-manager/02-trust-manager.yaml; do sleep 5; done'
  wait_until_pods_running "$CERT_MANAGER_NAMESPACE" || fail_test "Failed to install cert manager"

  kubectl apply -f "${root_dir}/test/config" || return $?
}
