package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/validator"
)

var (
	files       []string
	langs       []string
	maxParallel uint
	jsonOut     bool
	noColor     bool

	clrReset  = "\x1b[0m"
	clrRed    = "\x1b[31m"
	clrGreen  = "\x1b[32m"
	clrYellow = "\x1b[33m"
	clrCyan   = "\x1b[36m"
)

type fileOutcome struct {
	Idx        int                `json:"-"`
	Path       string             `json:"path"`
	Output     string             `json:"-"`
	Passed     int                `json:"passed"`
	Warned     int                `json:"warned"`
	Failed     int                `json:"failed"`
	Errored    int                `json:"errored"`
	HadOpErr   bool               `json:"had_op_err"`
	HadValFail bool               `json:"had_val_fail"`
	Summary    *validator.Summary `json:"summary,omitempty"`
}

type job struct {
	idx  int
	path string
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate one or multiple glossary files for structure, encoding and consistency",
	Long: `Run a full set of validation checks against one or multiple glossary CSV files.

The validator inspects file encoding, header correctness, column uniqueness,
language declarations, boolean flags, and term consistency. It can detect
malformed headers, unknown or missing language codes, duplicate entries, or
non-UTF8 input.

Each file is analyzed independently and reported with PASS / WARN / FAIL / ERROR status.

Examples:
  # Validate a single glossary file
  glossary-guard validate -f glossary.csv

  # Validate several files listed explicitly
  glossary-guard validate -f file1.csv -f file2.csv

  # Validate multiple files with declared languages
  glossary-guard validate -f glossary.csv -l en -l de -l fr

  # Using comma-separated file list (equivalent)
  glossary-guard validate -f file1.csv,file2.csv -l en_US,de_DE

By default, all built-in checks are executed:
  - UTF-8 encoding
  - Known headers and optional language columns
  - Orphaned *_description columns
  - Duplicate headers
  - Duplicate terms (case-sensitive)
  - Invalid Y/N flags (casesensitive/translatable/forbidden)
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(files) == 0 {
			return fmt.Errorf("no files provided; use --files to specify one or more CSV files")
		}

		if !noColor && os.Getenv("NO_COLOR") != "" {
			noColor = true
		}

		langs = preprocessLangs(langs)

		var err error
		files, err = expandFiles(files)

		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		if len(checks.All) == 0 {
			fmt.Fprintln(os.Stderr, red("No checks registered. Nothing to run."))
			return fmt.Errorf("no checks to run")
		}

		critFlags := make(map[string]bool)
		crit, norm := checks.Split()

		for _, c := range crit {
			critFlags[c.Name()] = true
		}

		for _, c := range norm {
			if _, ok := critFlags[c.Name()]; !ok {
				critFlags[c.Name()] = false
			}
		}

		sep := strings.Repeat("─", 72)

		jobs := make(chan job)
		outcomes := make([]fileOutcome, len(files))

		var wg sync.WaitGroup

		if maxParallel < 1 {
			maxParallel = uint(runtime.GOMAXPROCS(0))
		}

		workers := min(int(maxParallel), len(files))
		workers = max(1, workers)

		wg.Add(workers)

		for w := 0; w < workers; w++ {
			go func() {
				defer wg.Done()
				for j := range jobs {
					oc := runOneFile(j.idx, j.path, critFlags, langs, sep)
					outcomes[j.idx] = oc
				}
			}()
		}

		go func() {
			for i, p := range files {
				jobs <- job{idx: i, path: p}
			}
			close(jobs)
		}()

		wg.Wait()

		return finalize(outcomes, len(files), start)
	},
}

func Init(root *cobra.Command) {
	validateCmd.Flags().StringSliceVarP(
		&files,
		"files",
		"f",
		nil,
		"Path(s) to glossary file(s) to validate (comma-separated or repeatable, supports globs)",
	)

	validateCmd.Flags().UintVar(
		&maxParallel,
		"parallel",
		uint(runtime.GOMAXPROCS(0)),
		"Maximum number of files to validate in parallel",
	)

	validateCmd.Flags().StringSliceVarP(
		&langs,
		"langs",
		"l",
		nil,
		"Language codes to expect in the header (e.g. en,fr,de or de_DE,pt-BR)",
	)

	validateCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output (also honored if NO_COLOR is set)")

	validateCmd.Flags().BoolVar(&jsonOut, "json", false, "Output results as JSON (machine-readable)")

	root.AddCommand(validateCmd)
}

func preprocessLangs(ls []string) []string {
	if len(ls) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(ls))
	var out []string

	for _, v := range ls {
		for part := range strings.SplitSeq(v, ",") {
			s := strings.TrimSpace(part)
			if s == "" {
				continue
			}

			if _, ok := seen[s]; ok {
				continue
			}

			seen[s] = struct{}{}
			out = append(out, s)
		}
	}

	sort.Strings(out)
	return out
}

func expandFiles(fs []string) ([]string, error) {
	seen := map[string]struct{}{}
	var out []string

	for _, f := range fs {
		for p := range strings.SplitSeq(f, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			if hasGlob(p) {
				matches, err := filepath.Glob(p)
				if err != nil {
					return nil, err
				}
				for _, m := range matches {
					info, err := os.Stat(m)
					if err == nil && !info.IsDir() {
						if _, ok := seen[m]; !ok {
							seen[m] = struct{}{}
							out = append(out, m)
						}
					}
				}
				continue
			}

			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no files matched the provided patterns")
	}
	return out, nil
}

func hasGlob(s string) bool {
	return strings.ContainsAny(s, "*?[]")
}

func finalize(outcomes []fileOutcome, filesCount int, start time.Time) error {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(outcomes); err != nil {
			fmt.Fprintln(os.Stderr, red(fmt.Sprintf("failed to encode json: %v", err)))
			return err
		}
		return aggregateReturnCode(outcomes)
	}

	hadOpErr, hadValFail, filesPassed, filesFailed, filesErrored := printAndAggregate(outcomes, filesCount, start)

	if hadOpErr {
		return fmt.Errorf("one or more files could not be validated due to an error")
	}
	if hadValFail {
		return validator.ErrValidationFailed
	}

	_ = filesPassed
	_ = filesFailed
	_ = filesErrored

	return nil
}

func printAndAggregate(outcomes []fileOutcome, filesCount int, start time.Time) (hadOpErr, hadValFail bool, filesPassed, filesFailed, filesErrored int) {
	var totalWarns int

	for _, oc := range outcomes {
		if oc.Output != "" {
			fmt.Print(oc.Output)
		}
		filesPassed += oc.Passed
		filesFailed += oc.Failed
		filesErrored += oc.Errored
		totalWarns += oc.Summary.Warn
		hadOpErr = hadOpErr || oc.HadOpErr
		hadValFail = hadValFail || oc.HadValFail
	}

	if filesCount > 1 {
		fmt.Println()
		fmt.Printf("Overall: %s passed, %s warning(s), %s failed, %s error(s)\n",
			green(fmt.Sprint(filesPassed)),
			yellow(fmt.Sprint(totalWarns)),
			red(fmt.Sprint(filesFailed)),
			red(fmt.Sprint(filesErrored)),
		)
	}

	fmt.Printf("\nTotal time: %v\n", time.Since(start).Round(time.Millisecond))
	return hadOpErr, hadValFail, filesPassed, filesFailed, filesErrored
}

func aggregateReturnCode(outcomes []fileOutcome) error {
	var hadOpErr, hadValFail bool
	for _, oc := range outcomes {
		hadOpErr = hadOpErr || oc.HadOpErr
		hadValFail = hadValFail || oc.HadValFail
	}
	if hadOpErr {
		return fmt.Errorf("one or more files could not be validated due to an error")
	}
	if hadValFail {
		return validator.ErrValidationFailed
	}
	return nil
}

func runOneFile(i int, path string, critFlags map[string]bool, langs []string, sep string) fileOutcome {
	var b strings.Builder
	if i > 0 {
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%s\n%s: %s\n%s\n\n", sep, cyan("Validating"), path, sep)

	oc := fileOutcome{Idx: i, Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(&b, "%s: %v\n%s\n", red("ERROR"), err, sep)
		oc.HadOpErr = true
		oc.Errored++
		oc.Output = b.String()
		return oc
	}

	sum, err := validator.Validate(data, path, langs)
	oc.Summary = &sum

	if err != nil && !errors.Is(err, validator.ErrValidationFailed) {
		fmt.Fprintf(&b, "%s: %v\n%s\n", red("ERROR"), err, sep)
		oc.HadOpErr = true
		oc.Errored++
		oc.Output = b.String()
		return oc
	}

	for _, r := range sum.Results {
		tag := "NORM"
		if critFlags[r.Name] {
			tag = "CRIT"
		}
		fmt.Fprintf(&b, "→ [%s] %s ... %s\n   %s\n",
			tag, r.Name, colorStatus(string(r.Status)), r.Message)
	}

	fmt.Fprintf(&b, "\nSummary for %s: %s passed, %s warning(s), %s failed, %s errors\n",
		path,
		green(fmt.Sprint(sum.Pass)),
		yellow(fmt.Sprint(sum.Warn)),
		red(fmt.Sprint(sum.Fail)),
		red(fmt.Sprint(sum.Error)),
	)

	if sum.EarlyExit {
		total := len(checks.All)
		skipped := 0
		if total > len(sum.Results) {
			skipped = total - len(sum.Results)
		}
		fmt.Fprintf(&b, "%s due to fail-fast in check %q (%s). Skipped %d remaining check(s).\n",
			red("Stopped early"),
			sum.EarlyCheck, string(sum.EarlyStatus), skipped)
	}

	if sum.Fail > 0 || sum.Error > 0 || (err != nil && errors.Is(err, validator.ErrValidationFailed)) {
		fmt.Fprintln(&b, red("Result: FAILED"))
		oc.Failed++
		oc.HadValFail = true
	} else {
		if sum.Warn > 0 {
			fmt.Fprintln(&b, yellow("Result: PASSED WITH WARNINGS"))
			oc.Warned++
		} else {
			fmt.Fprintln(&b, green("Result: PASSED"))
			oc.Passed++
		}
	}

	fmt.Fprintf(&b, "%s\n", sep)
	oc.Output = b.String()
	return oc
}

func green(s string) string {
	if noColor {
		return s
	}
	return clrGreen + s + clrReset
}

func red(s string) string {
	if noColor {
		return s
	}
	return clrRed + s + clrReset
}

func cyan(s string) string {
	if noColor {
		return s
	}
	return clrCyan + s + clrReset
}

func yellow(s string) string {
	if noColor {
		return s
	}
	return clrYellow + s + clrReset
}

func colorStatus(s string) string {
	switch s {
	case "PASS":
		return green(s)
	case "WARN":
		return yellow(s)
	default: // FAIL, ERROR, unknown
		return red(s)
	}
}
