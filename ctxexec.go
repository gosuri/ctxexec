// Package ctxexec provides helper functions for running context-aware external commands.
package ctxexec

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/net/context"
)

// StopFunc is the function that terminates a command
type StopFunc func(ctx context.Context, cmd *exec.Cmd) error

// Stopper wrapps the *exec.Cmd with a StopFunc
// It provides context-aware graceful termination helper functions.
type Stopper struct {
	// StopFunc is the function to call when stopping the command
	StopFunc
	*exec.Cmd // Cmd represents an external command being prepared or run
}

// NewStopper returns a new Stopper for the *exec.Cmd with a default StopFunc
func NewStopper(cmd *exec.Cmd) *Stopper {
	return &Stopper{Cmd: cmd, StopFunc: stopFunc}
}

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
//
// If the command fails to run or doesn't complete successfully, the
// error is of type *exec.ExitError, context.DeadlineExceeded,
// context.Canceled. Other error types may be returned for I/O problems.
func Run(ctx context.Context, cmd *exec.Cmd) error {
	return NewStopper(cmd).Run(ctx)
}

// Stop terminates commmand execution using a new Stopper
//
// The returned error is nil if the command stopped before
// the context was cancelled
//
// It gracefully waits for the command to finish termination
// before killing the process when the context is cancelled
func Stop(ctx context.Context, cmd *exec.Cmd) error {
	return NewStopper(cmd).Run(ctx)
}

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
//
// If the command fails to run or doesn't complete successfully, the
// error is of type *exec.ExitError, context.DeadlineExceeded,
// context.Canceled. Other error types may be returned for I/O problems.
func (c *Stopper) Run(ctx context.Context) error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait(ctx)
}

// Start starts the specified command but does not wait for it to complete.
//
// The Wait method will return the exit code and release associated resources
// once the command exits.
func (c *Stopper) Start() error {
	return c.Cmd.Start()
}

// Stop terminates the execution when the command is running.
//
// The returned error is nil if the command stopped before the context
// was cancelled
//
// It gracefully waits for the command to finish execution before killing
// it after a timeout.
func (c *Stopper) Stop(ctx context.Context) error {
	return c.StopFunc(ctx, c.Cmd)
}

// stopFunc is the default function used for terminating the command exectution
func stopFunc(ctx context.Context, cmd *exec.Cmd) error {
	// try graceful termination first
	cmd.Process.Signal(os.Interrupt)
	cmd.Process.Signal(syscall.SIGTERM)
	// wait for process to finish terminating, kill when context is cancelled
	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		return ctx.Err()
	default:
		if err := cmd.Wait(); err != nil {
			return err
		}
		return nil
	}
}

// Wait waits for the command to exit.
// It must have been started by Start.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
//
// If the command fails to run or doesn't complete successfully, the
// error is of type *ExitError. Other error types may be
// returned for I/O problems.
//
// If c.Stdin is not an *os.File, Wait also waits for the I/O loop
// copying from c.Stdin into the process's standard input
// to complete.
//
// Wait releases any resources associated with the Cmd.
func (c *Stopper) Wait(ctx context.Context) error {
	<-ctx.Done()
	c.Stop(ctx)
	if err := c.Cmd.Wait(); err != nil { // wait for the process to be killed
		return err
	}
	return ctx.Err()
}

// stopped returns true if the process stopped and created a process state
func (c *Stopper) stopped() bool {
	return c.Cmd.ProcessState != nil // ProcessState is created only after the process stop running
}
