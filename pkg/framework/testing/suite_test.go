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

package testing

import (
	"fmt"
	"testing"

	"knative.dev/reconciler-test/pkg/framework"
)

type Config struct {
	framework.BaseConfig
	Broker string
}

var config = Config{}

func TestMain(m *testing.M) {
	framework.NewSuite(m).Configure(&config).Run()
}

func TestUnwrapped(t *testing.T) {
	fmt.Println("broker is " + config.Broker)
}

func TestWrapped(t *testing.T) {
	framework.NewTest(t).Feature("Broker").Run(func(tc framework.TestContext) {
		fmt.Println("broker is " + config.Broker)
	})
}

func TestMust(t *testing.T) {
	framework.NewTest(t).
		Feature("BrokerFeature").
		Must("").
		Run(func(tc framework.TestContext) {
			fmt.Println("broker is " + config.Broker)
		})
}
