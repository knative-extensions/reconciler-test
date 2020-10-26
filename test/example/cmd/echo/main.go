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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kelseyhightower/envconfig"
	"knative.dev/reconciler-test/test/example/config/echo"
)

func main() {
	r := Echo{}
	if err := envconfig.Process("", &r); err != nil {
		if err := r.WriteTerminationMessage(echo.Output{Success: false, Message: err.Error()}); err != nil {
			fmt.Printf("failed to write termination message, %s.\n", err)
		}
	}

	if err := r.Do(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

type Echo struct {
	Echo string `envconfig:"ECHO" required:"true"`
}

func (e *Echo) Do() error {
	fmt.Printf("%s\n", e.Echo)

	if err := e.WriteTerminationMessage(echo.Output{Success: true, Message: e.Echo}); err != nil {
		fmt.Printf("failed to write termination message, %v.\n", err)
		return err
	}

	return nil
}

// WriteTerminationMessage writes the result into the termination log as json.
func (e *Echo) WriteTerminationMessage(result interface{}) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("/dev/termination-log", b, 0644)
}
