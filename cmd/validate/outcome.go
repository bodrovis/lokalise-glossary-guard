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

func fileReadErrorOutcome(i int, path string, b *strings.Builder, err error, sep string) fileOutcome {
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
