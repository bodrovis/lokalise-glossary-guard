package validate

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"fmt"
	"io"
	"time"
)

type finalizer struct {
	out     io.Writer
	errOut  io.Writer
	colors  colorizer
	jsonOut bool
}

func newFinalizer(
	out io.Writer,
	errOut io.Writer,
	jsonOut bool,
	colors colorizer,
) finalizer {
	return finalizer{
		out:     out,
		errOut:  errOut,
		colors:  colors,
		jsonOut: jsonOut,
	}
}

func (f finalizer) finalize(
	outcomes []fileOutcome,
	filesCount int,
	start time.Time,
) error {
	if f.jsonOut {
		return f.finalizeJSON(outcomes)
	}

	return f.finalizeText(outcomes, filesCount, start)
}

func (f finalizer) finalizeJSON(outcomes []fileOutcome) error {
	if err := json.MarshalWrite(
		f.out,
		outcomes,
		jsontext.WithIndent("  "),
	); err != nil {
		_, _ = fmt.Fprintln(
			f.errOut,
			f.colors.red("failed to encode json: "+err.Error()),
		)
		return err
	}

	if _, err := io.WriteString(f.out, "\n"); err != nil {
		_, _ = fmt.Fprintln(
			f.errOut,
			f.colors.red("failed to write json: "+err.Error()),
		)
		return err
	}

	return aggregateReturnCode(outcomes)
}

func (f finalizer) finalizeText(
	outcomes []fileOutcome,
	filesCount int,
	start time.Time,
) error {
	agg := f.printAndAggregate(
		outcomes,
		filesCount,
		start,
	)

	return aggregateError(agg)
}

func (f finalizer) printAndAggregate(
	outcomes []fileOutcome,
	filesCount int,
	start time.Time,
) aggregateResult {
	for _, oc := range outcomes {
		if oc.Output != "" {
			fmt.Fprint(f.out, oc.Output)
		}
	}

	agg := aggregateOutcomes(outcomes)

	if filesCount > 1 {
		fmt.Fprintln(f.out)

		fmt.Fprintf(
			f.out,
			"Overall: %s passed, %s warning(s), %s failed, %s error(s)\n",
			f.colors.green(fmt.Sprint(agg.Passed)),
			f.colors.yellow(fmt.Sprint(agg.Warns)),
			f.colors.red(fmt.Sprint(agg.Failed)),
			f.colors.red(fmt.Sprint(agg.Errored)),
		)
	}

	fmt.Fprintf(
		f.out,
		"\nTotal time: %v\n",
		time.Since(start).Round(time.Millisecond),
	)

	return agg
}
