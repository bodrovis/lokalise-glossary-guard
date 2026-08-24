package cmd

import (
	"github.com/bodrovis/lokalise-glossary-guard/cmd/validate"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "glossary-guard",
		Short: "Validate Lokalise glossary CSVs",
		Long: `glossary-guard validates CSV files before uploading them to Lokalise.

It checks UTF-8 encoding, header shape, optional language columns, duplicate headers/terms,
and Y/N flags catching the most common issues (wrong delimiter, missing term/description, etc.).`,
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		validate.NewCmd(),
		newVersionCmd(),
		newGenDocsCmd(),
	)

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Printf(
				"glossary-guard %s\ncommit: %s\nbuilt at: %s\n",
				version,
				commit,
				date,
			)
		},
	}
}

func newGenDocsCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "gendocs",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return generateDocs(cmd.Root(), "./docs")
		},
	}
}
