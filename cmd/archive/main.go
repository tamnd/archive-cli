// Command archive is a single-binary command line for the Internet Archive.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/tamnd/archive-cli/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cli.Root()
	// fang gives us styled help, errors, and shell completion for free, and
	// returns the underlying error unwrapped, so the coded-exit type set by the
	// cli package survives for the exit-code mapping below.
	if err := fang.Execute(ctx, root,
		fang.WithVersion(cli.Version),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	); err != nil {
		os.Exit(cli.ExitCode(err))
	}
}
