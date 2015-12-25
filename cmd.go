// Package ctxexec provides helper functions for running context-aware external commands.
package ctxexec

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/net/context"
)

// Cmd is a wrapper for exec.Cmd that provides context-aware graceful termination helper functions
type Cmd struct {
	// StopFunc is the function to call when stopping the command
	StopFunc func(ctx context.Context, cmd *exec.Cmd) error
	cmd      *exec.Cmd // Cmd represents an external command being prepared or run
}

// New returns a new Cmd
func New(cmd *exec.Cmd) *Cmd {
	return &Cmd{cmd: cmd}
}

// StopFunc is the default function used for terminating the command exectution
func StopFunc(ctx context.Context, cmd *exec.Cmd) error {
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

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
//
// If the command fails to run or doesn't complete successfully, the
// error is of type *exec.ExitError. Other error types may be
// returned for I/O problems.
func (c *Cmd) Run(ctx context.Context) error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait(ctx)
}

// Start starts the specified command but does not wait for it to complete.
//
// The Wait method will return the exit code and release associated resources
// once the command exits.
func (c *Cmd) Start() error {
	return c.cmd.Start()
}

// Stop terminates the execution when the command is running.
//
// It gracefully waits for the command to finish execution before killing it after a timeout.
func (c *Cmd) Stop(ctx context.Context) error {
	if c.StopFunc != nil { // stop using the user provided stop function
		return c.StopFunc(ctx, c.cmd)
	}
	return StopFunc(ctx, c.cmd)
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
func (c *Cmd) Wait(ctx context.Context) error {
	<-ctx.Done()
	c.Stop(ctx)
	if err := c.cmd.Wait(); err != nil { // wait for the process to be killed
		return err
	}
	return ctx.Err()
}

// stopped returns true if the process stopped and created a process state
func (c *Cmd) stopped() bool {
	return c.cmd.ProcessState != nil // ProcessState is created only after the process stop running
}
