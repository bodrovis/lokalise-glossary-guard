package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

func finalize(outcomes []fileOutcome, filesCount int, start time.Time, jsonOutput bool) error {
	if jsonOutput {
		return finalizeJSON(os.Stdout, os.Stderr, outcomes)
	}

	return finalizeText(outcomes, filesCount, start)
}

func finalizeJSON(out io.Writer, errOut io.Writer, outcomes []fileOutcome) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	if err := enc.Encode(outcomes); err != nil {
		_, _ = fmt.Fprintln(errOut, red("failed to encode json: "+err.Error()))
		return err
	}

	return aggregateReturnCode(outcomes)
}

func finalizeText(outcomes []fileOutcome, filesCount int, start time.Time) error {
	hadOpErr, hadValFail, _, _, _ := printAndAggregate(outcomes, filesCount, start)

	if hadOpErr {
		return fmt.Errorf("one or more files could not be validated due to an error")
	}

	if hadValFail {
		return fmt.Errorf("validation failed")
	}

	return nil
}
