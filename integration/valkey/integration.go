package valkey

import (
	"context"

	"github.com/mountayaapp/helix.go/errorstack"
	"github.com/mountayaapp/helix.go/integration"
)

/*
Ensure *connection complies to the integration.Integration type.
*/
var _ integration.Integration = (*connection)(nil)

/*
String returns the string representation of the Valkey integration.
*/
func (conn *connection) String() string {
	return identifier
}

/*
Start does nothing since the Valkey integration only exposes a client to communicate
with a Valkey database.
*/
func (conn *connection) Start(ctx context.Context) error {
	return nil
}

/*
Close tries to gracefully close the connection with the database.
*/
func (conn *connection) Close(ctx context.Context) error {
	conn.client.Close()

	return nil
}

/*
Status indicates if the integration is able to ping the Valkey database or not.
Returns `200` if connection is working, `503` otherwise.
*/
func (conn *connection) Status(ctx context.Context) (int, error) {
	stack := errorstack.New("Integration is not in a healthy state", errorstack.WithIntegration(identifier))

	err := conn.client.Do(ctx, conn.client.B().Ping().Build()).Error()
	if err != nil {
		stack.WithValidations(errorstack.Validation{
			Message: err.Error(),
		})

		return 503, stack
	}

	return 200, nil
}
