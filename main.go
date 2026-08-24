package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/bodrovis/lokalise-glossary-guard/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	os.Exit(run(
		ctx,
		os.Args[1:],
		os.Stdout,
		os.Stderr,
	))
}

func run(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	rootCmd := cmd.RootCmd()

	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}
