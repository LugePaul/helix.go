package temporal

import (
	"context"

	"go.temporal.io/sdk/client"
)

/*
iclient is the internal client used as Temporal client. It implements the Client
interface and allows to wrap the Temporal's client functions for automatic tracing
and error recording.
*/
type iclient struct {
	config *Config
	client client.Client
}

/*
Client exposes an opinionated way to interact with Temporal's client capabilities.
*/
type Client interface {
	executeWorkflow(ctx context.Context, opts client.StartWorkflowOptions, workflowType string, payload ...any) (client.WorkflowRun, error)
	createSchedule(ctx context.Context, opts client.ScheduleOptions) error
}

/*
executeWorkflow starts a workflow execution and return a WorkflowRun instance and
error.

It automatically handles tracing and error recording via interceptor.
*/
func (c *iclient) executeWorkflow(ctx context.Context, opts client.StartWorkflowOptions, workflowType string, payload ...any) (client.WorkflowRun, error) {
	return c.client.ExecuteWorkflow(ctx, opts, workflowType, payload...)
}

/*
createSchedule creates a new schedule of a workflow type.
*/
func (c *iclient) createSchedule(ctx context.Context, opts client.ScheduleOptions) error {
	_, err := c.client.ScheduleClient().Create(ctx, opts)

	return err
}
