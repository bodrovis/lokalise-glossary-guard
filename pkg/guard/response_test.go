package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/validator"
)

func TestResponseFromSummary_Error(t *testing.T) {
	t.Parallel()

	t.Run("stores non-context error", func(t *testing.T) {
		t.Parallel()

		resp := responseFromSummary(
			"glossary.csv",
			validator.Summary{},
			errors.New("validation exploded"),
		)

		if resp.Error != "validation exploded" {
			t.Fatalf(
				"Error = %q, want %q",
				resp.Error,
				"validation exploded",
			)
		}

		if resp.Status != StatusFailed {
			t.Fatalf(
				"Status = %q, want %q",
				resp.Status,
				StatusFailed,
			)
		}

		if !resp.Failed {
			t.Fatal("Failed = false, want true")
		}

		if !resp.Errored {
			t.Fatal("Errored = false, want true")
		}
	})

	t.Run("does not store context canceled", func(t *testing.T) {
		t.Parallel()

		resp := responseFromSummary(
			"glossary.csv",
			validator.Summary{},
			context.Canceled,
		)

		if resp.Error != "" {
			t.Fatalf(
				"Error = %q, want empty",
				resp.Error,
			)
		}
	})
}
