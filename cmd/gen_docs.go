package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func generateDocs(rootCmd *cobra.Command, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create docs directory %q: %w", dir, err)
	}

	if err := doc.GenMarkdownTree(rootCmd, dir); err != nil {
		return fmt.Errorf("generate Markdown docs in %q: %w", dir, err)
	}

	return nil
}
