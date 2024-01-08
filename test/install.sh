#!/usr/bin/env bash

source "$(dirname "$(dirname "$(readlink -f "${BASH_SOURCE[0]:-$0}")")")/test/e2e-common.sh"

test_setup
