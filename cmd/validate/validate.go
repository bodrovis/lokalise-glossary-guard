package validate

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func NewCmd() *cobra.Command {
	opts := defaultCommandOptions()

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate one or multiple glossary files; optionally apply auto-fixes to _fixed copies",
		Long: `Run all registered checks against one or multiple glossary CSV files.

Examples:
  # Validate a single file (no fixes)
  glossary-guard validate -f glossary.csv

  # Validate and attempt fixes (writes glossary_fixed.csv on change)
  glossary-guard validate -f glossary.csv --fix

  # Multiple files + explicit languages
  glossary-guard validate -f a.csv -f b.csv -l en -l de -l fr --fix

  # Glob + parallel workers
  glossary-guard validate -f "data/*.csv" --parallel 8
`,
		Args: cobra.NoArgs,

		PreRunE: func(_ *cobra.Command, _ []string) error {
			return validatePreRun(&opts)
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return runValidate(cmd.Context(), validateRunConfig{
				Files:       opts.files,
				Langs:       opts.langs,
				MaxParallel: opts.maxParallel,
				Options:     buildRunOptions(opts),
				JSONOut:     opts.jsonOut,
				NoColor:     opts.noColor,
				Out:         cmd.OutOrStdout(),
				ErrOut:      cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().StringSliceVarP(
		&opts.files,
		"files",
		"f",
		nil,
		"Path(s) to glossary file(s) (comma-separated or repeatable, supports globs)",
	)

	cmd.Flags().UintVar(
		&opts.maxParallel,
		"parallel",
		opts.maxParallel,
		"Maximum number of files to process in parallel",
	)

	cmd.Flags().StringSliceVarP(
		&opts.langs,
		"langs",
		"l",
		nil,
		"Language codes expected in header (e.g. en,fr,de or de_DE,pt-BR)",
	)

	cmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable colored output (also honored if NO_COLOR is set)")
	cmd.Flags().BoolVar(&opts.jsonOut, "json", false, "Output results as JSON (machine-readable)")

	cmd.Flags().BoolVar(&opts.doFix, "fix", false, "Attempt auto-fixes (writes *_fixed.csv on change)")
	cmd.Flags().BoolVar(&opts.hardFailOnErr, "hard-fail-on-error", false, "Exit non-zero when any check returns ERROR")
	cmd.Flags().BoolVar(&opts.rerunAfterFix, "rerun-after-fix", true, "Re-run validation after a successful fix")

	return cmd
}

func validatePreRun(opts *commandOptions) error {
	if len(opts.files) == 0 {
		return fmt.Errorf(
			"no files provided; use --files to specify one or more CSV files",
		)
	}

	if !opts.noColor && os.Getenv("NO_COLOR") != "" {
		opts.noColor = true
	}

	opts.langs = guard.PreprocessLangs(opts.langs)

	files, err := expandFiles(opts.files)
	if err != nil {
		return err
	}

	opts.files = files

	if len(checks.List()) == 0 {
		return fmt.Errorf("no checks registered; nothing to run")
	}

	return nil
}
