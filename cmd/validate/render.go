package validate

import (
	"fmt"
	"strings"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
)

func renderResult(b *strings.Builder, resp guard.ValidateResponse) {
	switch {
	case resp.Failed:
		fmt.Fprintln(b, red("Result: FAILED"))
	case resp.Warned:
		fmt.Fprintln(b, yellow("Result: PASSED WITH WARNINGS"))
	default:
		fmt.Fprintln(b, green("Result: PASSED"))
	}
}

func renderValidationReport(b *strings.Builder, path string, sum guard.Summary) {
	renderCheckOutcomes(b, sum)
	renderSummary(b, path, sum)
	renderEarlyExit(b, sum)
}

func renderSummary(b *strings.Builder, path string, sum guard.Summary) {
	fmt.Fprintf(b, "\nSummary for %s: %s passed, %s warning(s), %s failed, %s errors\n",
		path,
		green(fmt.Sprint(sum.Pass)),
		yellow(fmt.Sprint(sum.Warn)),
		red(fmt.Sprint(sum.Fail)),
		red(fmt.Sprint(sum.Errors)),
	)
}

func renderEarlyExit(b *strings.Builder, sum guard.Summary) {
	if !sum.EarlyExit {
		return
	}

	total := len(checks.List())
	skipped := 0
	if total > len(sum.Outcomes) {
		skipped = total - len(sum.Outcomes)
	}

	fmt.Fprintf(b, "%s due to fail-fast in check %q (%s). Skipped %d remaining check(s).\n",
		red("Stopped early"),
		sum.EarlyCheck,
		sum.EarlyStatus,
		skipped,
	)
}

func renderCheckOutcomes(b *strings.Builder, sum guard.Summary) {
	for _, o := range sum.Outcomes {
		tag := "NORM"
		if o.Critical {
			tag = "CRIT"
		}

		changed := ""
		if o.Changed {
			changed = " [changed]"
		}

		msg := oneLine(strings.TrimSpace(o.Message))
		if msg == "" {
			msg = "-"
		}

		note := oneLine(strings.TrimSpace(o.Note))
		if note != "" {
			msg = msg + " | note: " + note
		}

		fmt.Fprintf(b, "→ [%s] %s ... %s%s\n", tag, o.Name, colorStatus(o.Status), changed)
		fmt.Fprintf(b, "   %s\n", msg)
	}
}

func renderFileHeader(b *strings.Builder, i int, path string, sep string, opts checks.RunOptions) {
	if i > 0 {
		b.WriteByte('\n')
	}

	fmt.Fprintf(b, "%s\n%s: %s\n%s\n\n", sep, cyan("Validating"), path, sep)

	fmt.Fprintf(b, "Mode: FixMode=%v, RerunAfterFix=%v, HardFailOnErr=%v\n\n",
		opts.FixMode, opts.RerunAfterFix, opts.HardFailOnErr)
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.Join(strings.Fields(s), " ")
}
