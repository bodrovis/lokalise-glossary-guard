package validate

import (
	"context"
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
	_ "github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks/all"
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/validator"
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
		if err != nil {
			return err
		}
		if len(checks.List()) == 0 {
			fmt.Fprintln(os.Stderr, red("No checks registered. Nothing to run."))
			return fmt.Errorf("no checks to run")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		sep := strings.Repeat("─", 72)

		jobs := make(chan job)
		outcomes := make([]fileOutcome, len(files))

		if maxParallel < 1 {
			maxParallel = uint(runtime.GOMAXPROCS(0))
		}
		workers := min(int(maxParallel), len(files))
		workers = max(1, workers)

		var wg sync.WaitGroup
		wg.Add(workers)

		ctx := cmd.Context()
		opts := buildRunOptions()

		for w := 0; w < workers; w++ {
			go func() {
				defer wg.Done()
				for j := range jobs {
					outcomes[j.idx] = runOneFile(ctx, j.idx, j.path, langs, sep, opts)
				}
			}()
		}

		go func() {
			for i, p := range files {
				select {
				case <-ctx.Done():
					return
				case jobs <- job{idx: i, path: p}:
				}
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

func buildRunOptions() checks.RunOptions {
	fm := checks.FixNone
	if doFix {
		fm = checks.FixIfNotPass
	}
	return checks.RunOptions{
		FixMode:       fm,
		RerunAfterFix: rerunAfterFix,
		HardFailOnErr: hardFailOnErr,
	}
}

func preprocessLangs(ls []string) []string {
	if len(ls) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ls))
	out := make([]string, 0, len(ls))
	for _, v := range ls {
		for _, part := range strings.Split(v, ",") {
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
		for _, raw := range strings.Split(f, ",") {
			p := strings.TrimSpace(raw)
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

func hasGlob(s string) bool { return strings.ContainsAny(s, "*?[]") }

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
		return fmt.Errorf("validation failed")
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
		if oc.Summary != nil {
			totalWarns += oc.Summary.Warn
		}
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
		return fmt.Errorf("validation failed")
	}
	return nil
}

func runOneFile(ctx context.Context, i int, path string, langs []string, sep string, opts checks.RunOptions) fileOutcome {
	var b strings.Builder
	if i > 0 {
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%s\n%s: %s\n%s\n\n", sep, cyan("Validating"), path, sep)

	fmt.Fprintf(&b, "Mode: FixMode=%v, RerunAfterFix=%v, HardFailOnErr=%v\n\n",
		opts.FixMode, opts.RerunAfterFix, opts.HardFailOnErr)

	oc := fileOutcome{Idx: i, Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(&b, "%s: %v\n%s\n", red("ERROR"), err, sep)
		oc.HadOpErr = true
		oc.Errored++
		oc.Output = b.String()
		return oc
	}

	sum, verr := validator.Validate(ctx, path, data, langs, opts)
	oc.Summary = &sum

	// print check-by-check
	for _, o := range sum.Outcomes {
		tag := "NORM"
		if cu, ok := checks.Lookup(o.Result.Name); ok && cu.FailFast() {
			tag = "CRIT"
		}
		changed := ""
		if o.Final.DidChange {
			changed = " [changed]"
		}

		msg := oneLine(strings.TrimSpace(o.Result.Message))
		if msg == "" {
			msg = "-"
		}
		note := oneLine(strings.TrimSpace(o.Final.Note))
		if note != "" {
			msg = msg + " | note: " + note
		}

		fmt.Fprintf(&b, "→ [%s] %s ... %s%s\n", tag, o.Result.Name, colorStatus(string(o.Result.Status)), changed)
		fmt.Fprintf(&b, "   %s\n", msg)
	}

	fmt.Fprintf(&b, "\nSummary for %s: %s passed, %s warning(s), %s failed, %s errors\n",
		path,
		green(fmt.Sprint(sum.Pass)),
		yellow(fmt.Sprint(sum.Warn)),
		red(fmt.Sprint(sum.Fail)),
		red(fmt.Sprint(sum.Error)),
	)

	if sum.EarlyExit {
		total := len(checks.List())
		skipped := 0
		if total > len(sum.Outcomes) {
			skipped = total - len(sum.Outcomes)
		}
		fmt.Fprintf(&b, "%s due to fail-fast in check %q (%s). Skipped %d remaining check(s).\n",
			red("Stopped early"),
			sum.EarlyCheck, string(sum.EarlyStatus), skipped)
	}

	// write *_fixed if we applied fixes
	if opts.FixMode != checks.FixNone && sum.AppliedFixes {
		outPath := withFixedPostfix(sum.FinalPath)
		if writeErr := os.WriteFile(outPath, sum.FinalData, 0o644); writeErr != nil {
			fmt.Fprintf(&b, "%s writing fixed file: %v\n", red("ERROR"), writeErr)
			oc.HadOpErr = true
			oc.Errored++
		} else {
			fmt.Fprintf(&b, "%s wrote fixed file: %s (bytes=%d)\n", cyan("Info"), outPath, len(sum.FinalData))
		}
	}

	// overall result per file
	if sum.Fail > 0 || sum.Error > 0 || (verr != nil && !errors.Is(verr, context.Canceled)) {
		fmt.Fprintln(&b, red("Result: FAILED"))
		oc.Failed++
		oc.HadValFail = true
	} else if sum.Warn > 0 {
		fmt.Fprintln(&b, yellow("Result: PASSED WITH WARNINGS"))
		oc.Warned++
	} else {
		fmt.Fprintln(&b, green("Result: PASSED"))
		oc.Passed++
	}

	fmt.Fprintf(&b, "%s\n", sep)
	oc.Output = b.String()
	return oc
}

func withFixedPostfix(p string) string {
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	if strings.HasSuffix(base, "_fixed") {
		return base + ext
	}
	return base + "_fixed" + ext
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
	default:
		return red(s) // FAIL/ERROR
	}
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.Join(strings.Fields(s), " ")
}
