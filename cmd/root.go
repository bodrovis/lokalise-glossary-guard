package cmd

import (
	"fmt"
	"os"

	"github.com/bodrovis/lokalise-glossary-guard/cmd/validate"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var version = "dev"

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
	}

	validate.Init(rootCmd)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("glossary-guard %s\n", version)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:    "gendocs",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := generateDocs(rootCmd, "./docs"); err != nil {
				fmt.Fprintf(os.Stderr, "error generating docs: %v", err)
			}
		},
	})

	return rootCmd
}

func generateDocs(rootCmd *cobra.Command, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return doc.GenMarkdownTree(rootCmd, dir)
}
