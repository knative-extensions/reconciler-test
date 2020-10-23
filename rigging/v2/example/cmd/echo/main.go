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

package main

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"knative.dev/reconciler-test/rigging/pkg/runner"
)

func main() {
	r := Echo{}
	if err := envconfig.Process("", &r); err != nil {
		if err := r.WriteTerminationMessage(runner.Output{Success: false, Message: err.Error()}); err != nil {
			fmt.Printf("failed to write termination message, %s.\n", err)
		}
	}

	if err := r.Do(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

type Echo struct {
	runner.Runner
	Echo string `envconfig:"ECHO" required:"true"`
}

func (e *Echo) Do() error {
	fmt.Printf("%s\n", e.Echo)

	if err := e.WriteTerminationMessage(runner.Output{Success: true, Message: e.Echo}); err != nil {
		fmt.Printf("failed to write termination message, %v.\n", err)
		return err
	}

	return nil
}
