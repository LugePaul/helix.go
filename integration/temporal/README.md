# helix.go - Temporal integration

[![Go API reference](https://pkg.go.dev/badge/github.com/mountayaapp/helix.go.svg)](https://pkg.go.dev/github.com/mountayaapp/helix.go/integration/temporal)
[![Go Report Card](https://goreportcard.com/badge/github.com/mountayaapp/helix.go/integration/temporal)](https://goreportcard.com/report/github.com/mountayaapp/helix.go/integration/temporal)
[![GitHub Release](https://img.shields.io/github/v/release/mountayaapp/helix.go)](https://github.com/mountayaapp/helix.go/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

The Temporal integration provides an opinionated way to interact with Temporal
for durable executions.

## Trace attributes

The `temporal` integration sets the following trace attributes:
- `temporal.server.address`
- `temporal.namespace`
- `span.kind`

When applicable, these attributes can be set as well:
- `temporal.worker.taskqueue`
- `temporal.workflow.id`
- `temporal.workflow.run_id`
- `temporal.workflow.namespace`
- `temporal.workflow.type`
- `temporal.activity.id`
- `temporal.activity.type`
- `temporal.activity.attempt`

Example:
```
temporal.server.address: "temporal.mydomain.tld"
temporal.namespace: "default"
temporal.worker.taskqueue: "demo"
temporal.workflow.namespace: "default"
temporal.workflow.type: "hello_world"
span.kind: "internal"
```

## Usage

Install the Go module with:
```sh
$ go get github.com/mountayaapp/helix.go/integration/temporal
```

Define type-safe workflows and activities:
```go
import (
  "github.com/mountayaapp/helix.go/integration/temporal"
)

var MyWorkflow = temporal.NewWorkflow[
  types.WorkflowInput,
  types.WorkflowResult,
]("workflow-name")

var MyActivity = temporal.NewActivity[
  types.ActivityInput,
  types.ActivityResult,
]("activity-name")
```

### Worker

Register type-safe workflows and activities in a worker:
```go
import (
  "github.com/mountayaapp/helix.go/integration/temporal"
)

cfg := temporal.Config{
  Address:   "localhost:7233",
  Namespace: "default",
  Worker: temporal.ConfigWorker{
    Enabled:   true,
    TaskQueue: "demo",
  },
}

_, worker, err := temporal.Connect(cfg)
if err != nil {
  return err
}

MyWorkflow.Register(worker, TypeSafeFunction)
MyActivity.Register(worker, TypeSafeFunction)
```

### Execute workflows from a client

Execute type-safe workflows from a client:
```go
import (
  "github.com/mountayaapp/helix.go/integration/temporal"
  "github.com/mountayaapp/helix.go/service"
)

cfg := temporal.Config{
  Address:   "localhost:7233",
  Namespace: "default",
}

client, _, err := temporal.Connect(cfg)
if err != nil {
  return err
}

result, err := MyWorkflow.Execute(ctx, client, opts, TypeSafeInput)
if err != nil {
  // ...
}
```

### Execute activities from a workflow

Execute type-safe activities from a workflow:
```go
err := MyActivity.Execute(ctx, payload).GetResult(ctx, &result)
if err != nil {
  // ...
}
```
