#!/usr/bin/env bash

# Copyright 2022 The Knative Authors
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

rootdir="$(dirname "$(dirname "$(readlink -f "${BASH_SOURCE[0]:-$0}")")")"
readonly rootdir

source "${rootdir}/vendor/knative.dev/hack/e2e-tests.sh"

function knative_setup() {
  kubectl apply -f "${rootdir}/test/config"
}

initialize "$@"

set -Eeuo pipefail

go_test_e2e -timeout 10m ./test/...

success
