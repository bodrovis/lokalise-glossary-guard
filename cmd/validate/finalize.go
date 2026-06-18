package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func finalize(outcomes []fileOutcome, filesCount int, start time.Time, jsonOutput bool) error {
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		if err := enc.Encode(outcomes); err != nil {
			fmt.Fprintln(os.Stderr, red(fmt.Sprintf("failed to encode json: %v", err)))
			return err
		}

		return aggregateReturnCode(outcomes)
	}

	hadOpErr, hadValFail, _, _, _ := printAndAggregate(outcomes, filesCount, start)

	if hadOpErr {
		return fmt.Errorf("one or more files could not be validated due to an error")
	}

	if hadValFail {
		return fmt.Errorf("validation failed")
	}

	return nil
}
