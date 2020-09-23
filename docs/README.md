# Knative Test Framework

The Knative test framework satifies the following requirements:

- Build upon golang testing given the supporting tools and services that already
  exist
- Test Authors need the ability to run any subset of tests
- Test Authors need the ability to select setup phases (install all components,
  install only core, install core + ping source, run tests without installing…)
- Test Authors need the ability to invoke tests on an existing cluster and
  within an existing namespace (not supported yet)
- Conformance Authors need the ability to mark certain features with their
  maturity/state (alpha, beta, stable)
- Conformance Authors need the ability to mark functionality with different
  requirement levels
- Conformance Authors need the ability to compose, consume & reuse common
  configuration (ie. environment settings)
- Conformance Authors need the ability to provide defaults for various test
  options.
- Downstream implementors need the ability to consume and invoke upstream
  conformance tests.
- Downstream implementors need the ability to invoke a subset of tests based on
  requirements levels.
- Downstream implementors need the ability to invoke a subset of tests based on
  the maturity of a feature.
- Downstream implementors need the ability to consistently supply test options
  and override defaults.

## Getting Started

Start by creating a Test Suite:

```go
import (
    "testing"

    "knative.dev/reconciler-test/pkg/framework"
)

type Config struct {
  framework.BaseConfig
  BrokerName string `desc:"The name of the broker"`
}

var config = Config{}

func TestMain(m *testing.M) {
  framework.
    NewSuite(m).
    Configure(&config).
    Run()
}
```

`NewSuite` creates a Test Suite wrapper, [`Configure`](#configuration)s it and
`Run` Test Cases declared in the same package.

And then a Test Case:

```go
import (
    "testing"
    "fmt"

    "knative.dev/reconciler-test/pkg/framework"
)

func TestCase(t *testing.T) {
  framework.NewTest(t).
    Feature("Named Broker").
    Run(func(tc framework.TestContext) {
      fmt.Println("broker name is " + config.BrokerName)
    })
}
```

And then run it as usual:

```sh
go test -v ./... -broker-name=kafka
2020/09/23 09:58:30 maxprocs: Leaving GOMAXPROCS=12: CPU quota undefined
=== RUN   TestCase
broker name is kafka
```

## Phases

### Configuration

The configuration phase performs the following actions:

- It reads the configuration file named `config-test.yaml`. This file must be
  located in the same directory as the Test Suite or in a parent directory, up
  to the project root directory (defined as where go.mod is located).
- It parses command line options and overrides the configuration parameters as
  needed.

The framework comes with a set of configuration parameters. Here is a subset:

```yaml
# whether to build and publish images before running tests.
buildImages: false

# image repository.
# e.g. docker.io/user
# when ko, use $KO_DOCKER_REPO environment variable
imageRepository: ko

requirements:
  must: true
  may: true
  should: true
```

To override, let say `imageRepository` and `requirements.may`, run go test with
the following options:

```sh
go test -image-repository=us.icr.io/knative/testing -requirements-may=false
```

This will run the tests using an alternative image repository and will skip the
tests marked a `may`.

You can use `Configure` to initiate the configuration phase:

```go
var config = framework.BaseConfig{}

func TestMain(m *testing.M) {
  framework.
    NewSuite(m).
    Configure(&config).
    Run()
}
```

Downstream implementors can provide their own configuration by embedding
`BaseConfig`:

```go
type MyConfig struct {
  framework.BaseConfig
  BrokerName string `desc:"The name of the broker"`
}
```

### Test setup

The next step is to properly setup the cluster before running tests. The test
framework works by expressing cluster setup requirements via the `Require`
function. For instance:

```go
func TestMain(m *testing.M) {
  framework.
    NewSuite(m).
    Configure(&config).
    Require(eventing.Component).
    Run()
```

Use the `Require` function to indicate a particular [component](#components) is
required in order to run the tests. It performs the following actions:

- It checks the cluster already contains the component and that it matches the
  test suite configuration (e.g. make sure latest Eventing component is up and
  running). This action is only performed for cluster-scoped components.
  - When the component does not exist, it installs it (WIP: add a flag)
  - When the component exists but does not match the test suite configuration,
    the test fails.
  - Otherwise all good
- It registers container images. This action is only performed for
  non-cluster-scoped components.

### Running tests

Finally the last phase consists on running Test Cases.

```go
func TestMain(m *testing.M) {
  framework.
    ....
    Run()
```

Before running the Test Cases, the following actions are performed:

- Container images marked as required are built and publised, if needed.
- A new namespace is created.
  - Not Supported Yet: option for running tests in a single predefined
    namespace.
- A test context is created and configured with the previously created
  namespace.

After running the Test Cases, the following actions are performed:

- The namespace is deleted.

## Test Case

A Test Case is a small wrapper around a Go test:

```go
func TestCase(t *testing.T) {
  framework.NewTest(t).
    Feature("Named Broker").
    Run(func(tc framework.TestContext) {
      fmt.Println("broker name is " + config.BrokerName)
    })
}
```

### Feature

`Feature` is used to mark the test as testing a particular feature.

### Requirements

A Test Case can be marked as having a specific requirement level, either
`Must()`, `Should()` or `May()`.

The `requirements-<must|should|may>` can be used to enable or disabled tests
based on requirement levels.

### Maturity

A Test Case can be marked with their maturity state, either `Alpha()`, `Beta()`
or `Stable()`.

The `maturity-<alpha|beta|stable>` can be used to enable or disabled tests based
on maturity states.

### Sub Tests

Not yet supported.

## Components

A component packages a set of YAML specifications and container images into a
single, self-contained entity.
