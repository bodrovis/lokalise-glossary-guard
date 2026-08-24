package guard

import (
	"context"
	"errors"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/validator"
)

func responseFromSummary(
	path string,
	coreSummary validator.Summary,
	err error,
) ValidateResponse {
	resp := ValidateResponse{
		Path:    path,
		Summary: newSummary(coreSummary),
	}

	if err != nil && !errors.Is(err, context.Canceled) {
		resp.Error = err.Error()
	}

	applyStatus(&resp)
	applyFixData(&resp, coreSummary)

	return resp
}

func applyStatus(resp *ValidateResponse) {
	resp.Passed = false
	resp.Warned = false
	resp.Failed = false

	resp.Errored = resp.Summary.Errors > 0 || resp.Error != ""

	switch {
	case resp.Summary.Fail > 0 || resp.Errored:
		resp.Status = StatusFailed
		resp.Failed = true

	case resp.Summary.Warn > 0:
		resp.Status = StatusPassedWithWarnings
		resp.Warned = true

	default:
		resp.Status = StatusPassed
		resp.Passed = true
	}
}

func applyFixData(
	resp *ValidateResponse,
	coreSummary validator.Summary,
) {
	if !coreSummary.AppliedFixes {
		return
	}

	resp.Fixed = true
	resp.FixedData = coreSummary.FinalData
	resp.FixedText = string(coreSummary.FinalData)
}
