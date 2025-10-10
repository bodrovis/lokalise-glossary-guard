package validate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bodrovis/lokalise-glossary-guard/internal/checks"
	"github.com/bodrovis/lokalise-glossary-guard/pkg/validator"
)

var files []string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate one or multiple glossary files",
	Long: `Validate one or multiple glossary CSV files by running all built-in checks.

Example:
  glossary-guard validate -f glossary.csv
  glossary-guard validate -f file1.csv,file2.csv
  glossary-guard validate -f file1.csv -f file2.csv
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(files) == 0 {
			return fmt.Errorf("no files provided — use --files to specify one or more CSV files")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(checks.All) == 0 {
			fmt.Println("No checks registered. Nothing to run.")
			return fmt.Errorf("no checks to run")
		}

		sep := strings.Repeat("─", 72)

		var hadOpErr bool
		var hadValFail bool
		var filesPassed, filesFailed, filesErrored int

		for i, path := range files {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s\n", sep)
			fmt.Printf("Validating: %s\n", path)
			fmt.Printf("%s\n\n", sep)

			sum, err := validator.Validate(path)

			if err != nil && !errors.Is(err, validator.ErrValidationFailed) {
				fmt.Printf("ERROR: %v\n", err)
				fmt.Printf("%s\n", sep)
				hadOpErr = true
				filesErrored++
				continue
			}

			names := checks.Sorted()
			for i, r := range sum.Results {
				name := "Check"
				if i < len(names) {
					name = names[i].Name()
				}
				fmt.Printf("→ %s ... %s\n   %s\n", name, r.Status, r.Message)
			}

			fmt.Printf("\nSummary for %s: %d passed, %d failed, %d errors\n",
				path, sum.Pass, sum.Fail, sum.Error)

			if sum.EarlyExit {
				skipped := 0
				if total := len(names); total > len(sum.Results) {
					skipped = total - len(sum.Results)
				}
				fmt.Printf("Stopped early due to fail-fast in check %q (%s). Skipped %d remaining check(s).\n",
					sum.EarlyCheck, sum.EarlyStatus, skipped)
			}

			if (err != nil && errors.Is(err, validator.ErrValidationFailed)) || sum.Fail > 0 || sum.Error > 0 {
				fmt.Println("Result: FAILED")
				filesFailed++
				hadValFail = true
			} else {
				fmt.Println("Result: PASSED")
				filesPassed++
			}

			fmt.Printf("%s\n", sep)
		}

		if len(files) > 1 {
			fmt.Println()
			fmt.Printf("Overall: %d passed, %d failed, %d error(s)\n",
				filesPassed, filesFailed, filesErrored)
		}

		if hadOpErr {
			return fmt.Errorf("one or more files could not be validated due to an error")
		}
		if hadValFail {
			return validator.ErrValidationFailed
		}
		return nil
	},
}

func Init(root *cobra.Command) {
	validateCmd.Flags().StringSliceVarP(&files, "files", "f", nil, "Path(s) to glossary file(s) to validate (comma-separated or repeatable)")
	root.AddCommand(validateCmd)
}
