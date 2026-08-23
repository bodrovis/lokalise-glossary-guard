package validate

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
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
	if err := json.MarshalWrite(
		out,
		outcomes,
		jsontext.WithIndent("  "),
	); err != nil {
		_, _ = fmt.Fprintln(errOut, red("failed to encode json: "+err.Error()))
		return err
	}

	if _, err := io.WriteString(out, "\n"); err != nil {
		_, _ = fmt.Fprintln(errOut, red("failed to write json: "+err.Error()))
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
