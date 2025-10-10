package cmd

import (
	"github.com/bodrovis/lokalise-glossary-guard/cmd/validate"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "glossary-guard",
		Short:         "CLI tool for validating glossary CSV files for Lokalise",
		Long:          `glossary-guard is a small CLI utility that validates CSV files before uploading them as glossaries to Lokalise.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	validate.Init(rootCmd)
	return rootCmd
}
