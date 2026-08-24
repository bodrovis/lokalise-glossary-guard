package validate

import (
	"fmt"
	"strings"

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
	HadOpErr   bool           `json:"-"`
	HadValFail bool           `json:"-"`
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

func (cfg fileRunConfig) opErrorOutcome(
	i int,
	path string,
	b *strings.Builder,
	err error,
) fileOutcome {
	fmt.Fprintf(
		b,
		"%s: %v\n%s\n",
		cfg.Colors.red("ERROR"),
		err,
		cfg.Separator,
	)

	return fileOutcome{
		Idx:      i,
		Path:     path,
		Errored:  1,
		HadOpErr: true,
		Output:   b.String(),
	}
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

func aggregateError(agg aggregateResult) error {
	if agg.HadOpErr {
		return fmt.Errorf(
			"one or more files could not be validated due to an error",
		)
	}

	if agg.HadValFail {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func aggregateReturnCode(outcomes []fileOutcome) error {
	return aggregateError(aggregateOutcomes(outcomes))
}
