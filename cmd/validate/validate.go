package validate

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

var (
	files       []string
	langs       []string
	maxParallel uint
	jsonOut     bool
	noColor     bool

	doFix         bool
	hardFailOnErr bool
	rerunAfterFix bool
)

var validateCmd = &cobra.Command{
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
	PreRunE: validatePreRun,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runValidate(cmd.Context(), validateRunConfig{
			Files:       files,
			Langs:       langs,
			MaxParallel: maxParallel,
			Options:     buildRunOptions(),
			JSONOut:     jsonOut,
		})
	},
}

func Init(root *cobra.Command) {
	validateCmd.Flags().StringSliceVarP(
		&files,
		"files",
		"f",
		nil,
		"Path(s) to glossary file(s) (comma-separated or repeatable, supports globs)",
	)

	validateCmd.Flags().UintVar(
		&maxParallel,
		"parallel",
		uint(runtime.GOMAXPROCS(0)),
		"Maximum number of files to process in parallel",
	)

	validateCmd.Flags().StringSliceVarP(
		&langs,
		"langs",
		"l",
		nil,
		"Language codes expected in header (e.g. en,fr,de or de_DE,pt-BR)",
	)

	validateCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output (also honored if NO_COLOR is set)")
	validateCmd.Flags().BoolVar(&jsonOut, "json", false, "Output results as JSON (machine-readable)")

	validateCmd.Flags().BoolVar(&doFix, "fix", false, "Attempt auto-fixes (writes *_fixed.csv on change)")
	validateCmd.Flags().BoolVar(&hardFailOnErr, "hard-fail-on-error", false, "Exit non-zero when any check returns ERROR")
	validateCmd.Flags().BoolVar(&rerunAfterFix, "rerun-after-fix", true, "Re-run validation after a successful fix")

	root.AddCommand(validateCmd)
}

func validatePreRun(cmd *cobra.Command, args []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided; use --files to specify one or more CSV files")
	}

	if !noColor && os.Getenv("NO_COLOR") != "" {
		noColor = true
	}

	langs = guard.PreprocessLangs(langs)

	var err error
	files, err = expandFiles(files)
	if err != nil {
		return err
	}

	if len(checks.List()) == 0 {
		fmt.Fprintln(os.Stderr, red("No checks registered. Nothing to run."))
		return fmt.Errorf("no checks to run")
	}

	return nil
}
