package validate

import (
	"fmt"
	"strings"
	"time"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

type fileOutcome struct {
	Idx        int            `json:"-"`
	Path       string         `json:"path"`
	Output     string         `json:"-"`
	Passed     int            `json:"passed"`
	Warned     int            `json:"warned"`
	Failed     int            `json:"failed"`
	Errored    int            `json:"errored"`
	HadOpErr   bool           `json:"had_op_err"`
	HadValFail bool           `json:"had_val_fail"`
	Summary    *guard.Summary `json:"summary,omitempty"`
}

func applyGuardResponse(oc *fileOutcome, resp guard.ValidateResponse) {
	oc.Summary = &resp.Summary

	if resp.Errored {
		oc.Errored = 1
	}

	switch {
	case resp.Passed:
		oc.Passed = 1
	case resp.Warned:
		oc.Warned = 1
	case resp.Failed:
		oc.Failed = 1
		oc.HadValFail = true
	}
}

func fileOpErrorOutcome(i int, path string, b *strings.Builder, err error, sep string) fileOutcome {
	fmt.Fprintf(b, "%s: %v\n%s\n", red("ERROR"), err, sep)

	return fileOutcome{
		Idx:      i,
		Path:     path,
		Errored:  1,
		HadOpErr: true,
		Output:   b.String(),
	}
}

func aggregateReturnCode(outcomes []fileOutcome) error {
	agg := aggregateOutcomes(outcomes)

	if agg.HadOpErr {
		return fmt.Errorf("one or more files could not be validated due to an error")
	}
	if agg.HadValFail {
		return fmt.Errorf("validation failed")
	}
	return nil
}

func printAndAggregate(outcomes []fileOutcome, filesCount int, start time.Time) (
	hadOpErr, hadValFail bool,
	filesPassed, filesFailed, filesErrored int,
) {
	for _, oc := range outcomes {
		if oc.Output != "" {
			fmt.Print(oc.Output)
		}
	}

	agg := aggregateOutcomes(outcomes)

	if filesCount > 1 {
		fmt.Println()
		fmt.Printf("Overall: %s passed, %s warning(s), %s failed, %s error(s)\n",
			green(fmt.Sprint(agg.Passed)),
			yellow(fmt.Sprint(agg.Warns)),
			red(fmt.Sprint(agg.Failed)),
			red(fmt.Sprint(agg.Errored)),
		)
	}

	fmt.Printf("\nTotal time: %v\n", time.Since(start).Round(time.Millisecond))

	return agg.HadOpErr, agg.HadValFail, agg.Passed, agg.Failed, agg.Errored
}

type aggregateResult struct {
	HadOpErr   bool
	HadValFail bool
	Passed     int
	Warns      int
	Failed     int
	Errored    int
}

func aggregateOutcomes(outcomes []fileOutcome) aggregateResult {
	var agg aggregateResult

	for _, oc := range outcomes {
		agg.Passed += oc.Passed
		agg.Failed += oc.Failed
		agg.Errored += oc.Errored

		if oc.Summary != nil {
			agg.Warns += oc.Summary.Warn
		}

		agg.HadOpErr = agg.HadOpErr || oc.HadOpErr
		agg.HadValFail = agg.HadValFail || oc.HadValFail
	}

	return agg
}
