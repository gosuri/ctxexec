package ctxexec

import (
	"os/exec"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestWait(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	run := `trap "echo intr; exit 0" SIGINT SIGTERM; echo running sleep 1; exit 0`
	c := New(exec.Command("bash", "-c", run))
	c.Start()
	c.Wait(ctx)
	if !c.Cmd.ProcessState.Success() {
		t.Fatalf("process failed to exit successfully. %+v", c.Cmd.ProcessState)
	}
}

func TestWait_Kill(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	run := `trap "echo ignoring" SIGINT; while true; do echo running; sleep 1; done`
	c := New(exec.Command("bash", "-c", run))
	c.Start()
	c.Wait(ctx)
	if !c.stopped() {
		t.Fatal("expected stop")
	}
}

func TestStop(t *testing.T) {
	run := `trap "echo intr; exit 0" SIGINT SIGTERM; while true; do echo running; sleep 1; done`
	c := New(exec.Command("bash", "-c", run))
	c.Start()
	time.Sleep(time.Second)
	c.Stop(context.Background())
	c.Cmd.Wait()
	if !c.Cmd.ProcessState.Success() {
		t.Fatalf("process failed to exit successfully. %+v", c.Cmd.ProcessState)
	}
}

func TestStop_NoStart(t *testing.T) {
	run := `trap "echo intr; exit 0" SIGINT SIGTERM; while true; do echo running; sleep 1; done`
	c := New(exec.Command("bash", "-c", run))
	c.Stop(context.Background())
	c.Cmd.Wait()
	if c.Cmd.ProcessState != nil {
		t.Fatalf("process failed to exit successfully. %+v", c.Cmd.ProcessState)
	}
}
