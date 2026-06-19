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
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	validate.Init(rootCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("glossary-guard %s\ncommit: %s\nbuilt at: %s\n", version, commit, date)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:    "gendocs",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateDocs(rootCmd, "./docs")
		},
	})

	return rootCmd
}
