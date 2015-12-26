package ctxexec_test

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/gosuri/ctxexec"
	"golang.org/x/net/context"
)

func ExampleRun() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2) // kill after 2 secs
	defer cancel()                                                          // cancel when command main exits
	cmd := exec.Command("sh", "-c", `while true; sleep 1; done`)            // run forever
	if err := ctxexec.Run(ctx, cmd); err != nil {
		fmt.Printf("%T", err)
	}
	// Output: *exec.ExitError
}
