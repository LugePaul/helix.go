package service

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mountayaapp/helix.go/errorstack"
	"github.com/mountayaapp/helix.go/integration"
	"github.com/mountayaapp/helix.go/internal/logger"
	"github.com/mountayaapp/helix.go/internal/tracer"
)

/*
svc is the service being run at the moment. Only one service should be running
at a time. This is why end-users can't create multiple services in a single Go
application.
*/
var svc = new(service)

/*
service holds some information for the service running.
*/
type service struct {

	// mutex allows to lock/unlock access to the service when necessary.
	mutex sync.Mutex

	// isInitialized informs if the service has already been initialized. In other
	// words this informs if the Init() function has already been called and returned
	// with no error.
	isInitialized bool

	// isStopped informs if the service has already been stopped. In other words
	// this informs if the Stop() function has already been called and returned
	// with no error.
	isStopped bool

	// integrations is the list of integrations attached to the service.
	integrations []integration.Integration
}

/*
Start initializes the helix service, and starts each integration attached by
executing their Start function. This returns as soon as an interrupting signal
is catched or when an integration returns an error while starting it.
*/
func Start(ctx context.Context) error {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	stack := errorstack.New("Failed to initialize the service")
	if svc.isInitialized {
		stack.WithValidations(errorstack.Validation{
			Message: "Service has already been initialized",
		})

		return stack
	}

	if svc.isStopped {
		stack.WithValidations(errorstack.Validation{
			Message: "Cannot initialize a stopped service",
		})

		return stack
	}

	// Create a channel for receiving interrupting signals, and another one for
	// catching integration errors. The function will then return as soon as one
	// of the channel receives a value.
	done := make(chan os.Signal, 1)
	failed := make(chan error, 1)

	// Listen for interrupting signals.
	go func() {
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-done
	}()

	// For each integration attached, execute its Start function. If an error is
	// encountered, send the error as a child error to the channel.
	for _, inte := range svc.integrations {
		go func() {
			err := inte.Start(ctx)
			if err != nil {
				failed <- stack.WithChildren(err)
			}
		}()
	}

	// Return as soon as an interrupting signal is catched or when an integration
	// returns an error while starting it.
	svc.isInitialized = true
	select {
	case <-done:
		return nil
	case <-failed:
		return stack
	}
}

/*
Stop tries to gracefully close connections with all integrations. It then tries
to drain/close the tracer and logger.
*/
func Stop(ctx context.Context) error {
	svc.mutex.Lock()
	defer svc.mutex.Unlock()

	stack := errorstack.New("Failed to gracefully close service's connections")
	if !svc.isInitialized {
		stack.WithValidations(errorstack.Validation{
			Message: "Service must first be initialized",
		})

		return stack
	}

	if svc.isStopped {
		stack.WithValidations(errorstack.Validation{
			Message: "Service has already been stopped",
		})

		return stack
	}

	var wg sync.WaitGroup
	for _, inte := range svc.integrations {
		wg.Go(func() {
			err := inte.Close(ctx)
			if err != nil {
				stack.WithChildren(err)
			}
		})
	}

	wg.Wait()
	if stack.HasChildren() {
		return stack
	}

	if tracer.Exporter() != nil {
		if err := tracer.Exporter().Shutdown(ctx); err != nil {
			stack.WithChildren(&errorstack.Error{
				Message: "Failed to gracefully drain/close tracer",
				Validations: []errorstack.Validation{
					{
						Message: err.Error(),
					},
				},
			})
		}
	}

	// Ignore if the error is ENOTTY, as explained in this comment on GitHub:
	// https://github.com/uber-go/zap/issues/991#issuecomment-962098428.
	if logger.Logger() != nil {
		if err := logger.Logger().Sync(); err != nil {
			if !errors.Is(err, syscall.ENOTTY) {
				stack.WithChildren(&errorstack.Error{
					Message: "Failed to gracefully drain/close logger",
					Validations: []errorstack.Validation{
						{
							Message: err.Error(),
						},
					},
				})
			}
		}
	}

	if stack.HasChildren() {
		return stack
	}

	svc.isStopped = true
	return nil
}
