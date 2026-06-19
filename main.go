package main

import (
	"fmt"
	"io"
	"os"

	"github.com/bodrovis/lokalise-glossary-guard/cmd"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}
