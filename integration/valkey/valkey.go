package valkey

import (
	"context"
	"fmt"

	"github.com/mountayaapp/helix.go/errorstack"
	"github.com/mountayaapp/helix.go/service"
	"github.com/mountayaapp/helix.go/telemetry/trace"

	"github.com/valkey-io/valkey-go"
)

/*
Valkey exposes an opinionated way to interact with Valkey, by bringing automatic
distributed tracing as well as error recording within traces.
*/
type Valkey interface {
	Get(ctx context.Context, key string, opts *OptionsGet) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, opts *OptionsSet) error
}

/*
connection represents the valkey integration. It respects the integration.Integration
and Valkey interfaces.
*/
type connection struct {

	// config holds the Config initially passed when creating a new Valkey client.
	config *Config

	// client is the connection made with the Valkey client.
	client valkey.Client
}

/*
Connect tries to create a Valkey client given the Config. Returns an error if
Config is not valid or if the initialization failed.
*/
func Connect(cfg Config) (Valkey, error) {

	// No need to continue if Config is not valid.
	err := cfg.sanitize()
	if err != nil {
		return nil, err
	}

	// Start to build an error stack, so we can add validations as we go.
	stack := errorstack.New("Failed to initialize integration", errorstack.WithIntegration(identifier))
	conn := &connection{
		config: &cfg,
	}

	// Set the default Valkey config.
	var opts = valkey.ClientOption{
		InitAddress: []string{cfg.Address},
		Username:    cfg.User,
		Password:    cfg.Password,
	}

	// Set TLS options only if enabled in Config.
	if cfg.TLS.Enabled {
		var validations []errorstack.Validation

		opts.TLSConfig, validations = cfg.TLS.ToStandardTLS()
		if len(validations) > 0 {
			stack.WithValidations(validations...)
		}
	}

	// Try to connect to the Valkey database.
	conn.client, err = valkey.NewClient(opts)
	if err != nil {
		stack.WithValidations(errorstack.Validation{
			Message: err.Error(),
		})
	}

	// Stop here if error validations were encountered.
	if stack.HasValidations() {
		return nil, stack
	}

	// Try to attach the integration to the service.
	if err := service.Attach(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

/*
Get reads the value at key and returns its byte representation.

It automatically handles tracing and error recording.
*/
func (conn *connection) Get(ctx context.Context, key string, opts *OptionsGet) ([]byte, error) {
	ctx, span := trace.Start(ctx, trace.SpanKindClient, fmt.Sprintf("%s: Get", humanized))
	defer span.End()

	cmd := conn.client.B().Get().Key(key)

	value, err := conn.client.Do(ctx, cmd.Build()).AsBytes()
	if err != nil {
		if opts != nil && opts.ErrorRecordOnNotFound {
			span.RecordError("failed to get key", err)
		}
	}

	setKeyAttributes(span, key)

	return value, err
}

/*
Set writes bytes representation of the value, with some optional options.

It automatically handles tracing and error recording.
*/
func (conn *connection) Set(ctx context.Context, key string, value []byte, opts *OptionsSet) error {
	ctx, span := trace.Start(ctx, trace.SpanKindClient, fmt.Sprintf("%s: Set", humanized))
	defer span.End()

	cmd := conn.client.B().Set().Key(key).Value(string(value))
	if opts != nil && opts.TTL > 0 {
		cmd.Ex(opts.TTL)
	}

	err := conn.client.Do(ctx, cmd.Build()).Error()
	if err != nil {
		span.RecordError("failed to set key", err)
	}

	setKeyAttributes(span, key)

	return err
}
