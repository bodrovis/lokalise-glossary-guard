package validate

import (
	"fmt"
	"strings"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

type fileRenderer struct {
	out       *strings.Builder
	separator string
	options   checks.RunOptions
	colors    colorizer
}

func newFileRenderer(
	out *strings.Builder,
	cfg fileRunConfig,
) fileRenderer {
	return fileRenderer{
		out:       out,
		separator: cfg.Separator,
		options:   cfg.Options,
		colors:    cfg.Colors,
	}
}

func (r fileRenderer) result(resp guard.ValidateResponse) {
	switch {
	case resp.Failed:
		fmt.Fprintln(
			r.out,
			r.colors.red("Result: FAILED"),
		)

	case resp.Warned:
		fmt.Fprintln(
			r.out,
			r.colors.yellow("Result: PASSED WITH WARNINGS"),
		)

	default:
		fmt.Fprintln(
			r.out,
			r.colors.green("Result: PASSED"),
		)
	}
}

func (r fileRenderer) validationReport(
	path string,
	sum guard.Summary,
) {
	r.checkOutcomes(sum)
	r.summary(path, sum)
	r.earlyExit(sum)
}

func (r fileRenderer) summary(
	path string,
	sum guard.Summary,
) {
	fmt.Fprintf(
		r.out,
		"\nSummary for %s: %s passed, %s warning(s), %s failed, %s errors\n",
		path,
		r.colors.green(fmt.Sprint(sum.Pass)),
		r.colors.yellow(fmt.Sprint(sum.Warn)),
		r.colors.red(fmt.Sprint(sum.Fail)),
		r.colors.red(fmt.Sprint(sum.Errors)),
	)
}

func (r fileRenderer) earlyExit(sum guard.Summary) {
	if !sum.EarlyExit {
		return
	}

	skipped := max(
		0,
		len(checks.List())-len(sum.Outcomes),
	)

	fmt.Fprintf(
		r.out,
		"%s due to fail-fast in check %q (%s). Skipped %d remaining check(s).\n",
		r.colors.red("Stopped early"),
		sum.EarlyCheck,
		sum.EarlyStatus,
		skipped,
	)
}

func (r fileRenderer) checkOutcomes(sum guard.Summary) {
	for _, outcome := range sum.Outcomes {
		tag := "NORM"
		if outcome.Critical {
			tag = "CRIT"
		}

		changed := ""
		if outcome.Changed {
			changed = " [changed]"
		}

		msg := oneLine(outcome.Message)
		if msg == "" {
			msg = "-"
		}

		if note := oneLine(outcome.Note); note != "" {
			msg += " | note: " + note
		}

		fmt.Fprintf(
			r.out,
			"→ [%s] %s ... %s%s\n",
			tag,
			outcome.Name,
			r.colors.status(outcome.Status),
			changed,
		)

		fmt.Fprintf(r.out, "   %s\n", msg)
	}
}

func (r fileRenderer) fileHeader(
	index int,
	path string,
) {
	if index > 0 {
		r.out.WriteByte('\n')
	}

	fmt.Fprintf(
		r.out,
		"%s\n%s: %s\n%s\n\n",
		r.separator,
		r.colors.cyan("Validating"),
		path,
		r.separator,
	)

	fmt.Fprintf(
		r.out,
		"Mode: FixMode=%v, RerunAfterFix=%v, HardFailOnErr=%v\n\n",
		r.options.FixMode,
		r.options.RerunAfterFix,
		r.options.HardFailOnErr,
	)
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
