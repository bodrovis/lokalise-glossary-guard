package main

import (
	"fmt"
	"os"

	"github.com/bodrovis/lokalise-glossary-guard/cmd"
)

func main() {
	rootCmd := cmd.RootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "command failed: %v\n", err)
		os.Exit(1)
	}
}
