package guard

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	_ "github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks/all"
	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/validator"
)

type ValidateRequest struct {
	Path string `json:"path,omitempty"`

	// Data is for Go callers: CLI, tests, server-side integrations.
	Data []byte `json:"-"`

	// Text is for JSON/WASM/browser callers.
	// It uses "data" in JSON because that's what the browser API naturally sends.
	Text string `json:"data,omitempty"`

	Langs []string `json:"langs,omitempty"`

	Fix           bool  `json:"fix,omitempty"`
	RerunAfterFix *bool `json:"rerun_after_fix,omitempty"`
	HardFailOnErr bool  `json:"hard_fail_on_error,omitempty"`
}

type ValidateStatus string

const (
	StatusPassed             ValidateStatus = "passed"
	StatusPassedWithWarnings ValidateStatus = "passed_with_warnings"
	StatusFailed             ValidateStatus = "failed"
)

type ValidateResponse struct {
	Path   string         `json:"path,omitempty"`
	Status ValidateStatus `json:"status"`

	Passed  bool `json:"passed"`
	Warned  bool `json:"warned"`
	Failed  bool `json:"failed"`
	Errored bool `json:"errored"`

	Error string `json:"error,omitempty"`

	Fixed     bool   `json:"fixed"`
	FixedText string `json:"fixed_text,omitempty"`

	Summary Summary `json:"summary"`

	// FixedData is for Go callers. JSON/browser callers should use FixedText.
	FixedData []byte `json:"-"`
}

type Summary struct {
	FilePath string `json:"file_path"`

	Pass   int `json:"pass"`
	Warn   int `json:"warn"`
	Fail   int `json:"fail"`
	Errors int `json:"errors"`

	AppliedFixes bool `json:"applied_fixes"`

	EarlyExit   bool   `json:"early_exit"`
	EarlyCheck  string `json:"early_check,omitempty"`
	EarlyStatus string `json:"early_status,omitempty"`

	FinalPath string `json:"final_path,omitempty"`

	Outcomes []Outcome `json:"outcomes"`
}

type Outcome struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`

	Critical bool `json:"critical"`
	Changed  bool `json:"changed"`

	Note string `json:"note,omitempty"`
}

func ValidateBytesJSON(ctx context.Context, req ValidateRequest) ([]byte, error) {
	resp, err := ValidateBytes(ctx, req)

	if err != nil && errors.Is(err, context.Canceled) {
		return nil, err
	}

	return json.Marshal(resp)
}

func ValidateBytes(ctx context.Context, req ValidateRequest) (ValidateResponse, error) {
	opts := checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: rerunAfterFix(req),
		HardFailOnErr: req.HardFailOnErr,
	}

	if req.Fix {
		opts.FixMode = checks.FixIfNotPass
	}

	langs := PreprocessLangs(req.Langs)

	coreSummary, err := validator.Validate(ctx, req.Path, requestData(req), langs, opts)
	summary := newSummary(coreSummary)

	resp := ValidateResponse{
		Path:    req.Path,
		Summary: summary,
	}

	if err != nil && !errors.Is(err, context.Canceled) {
		resp.Error = err.Error()
	}

	switch {
	case summary.Fail > 0 || summary.Errors > 0 || resp.Error != "":
		resp.Status = StatusFailed
		resp.Failed = true
	case summary.Warn > 0:
		resp.Status = StatusPassedWithWarnings
		resp.Warned = true
	default:
		resp.Status = StatusPassed
		resp.Passed = true
	}

	if summary.Errors > 0 || resp.Error != "" {
		resp.Errored = true
	}

	if coreSummary.AppliedFixes {
		resp.Fixed = true
		resp.FixedData = coreSummary.FinalData
		resp.FixedText = string(coreSummary.FinalData)
	}

	return resp, err
}

func requestData(req ValidateRequest) []byte {
	if req.Data != nil {
		return req.Data
	}

	return []byte(req.Text)
}

func rerunAfterFix(req ValidateRequest) bool {
	if req.RerunAfterFix == nil {
		return true
	}

	return *req.RerunAfterFix
}

func newSummary(sum validator.Summary) Summary {
	outcomes := make([]Outcome, 0, len(sum.Outcomes))

	for _, o := range sum.Outcomes {
		outcomes = append(outcomes, Outcome{
			Name:     o.Result.Name,
			Status:   string(o.Result.Status),
			Message:  o.Result.Message,
			Critical: isCriticalCheck(o.Result.Name),
			Changed:  o.Final.DidChange,
			Note:     o.Final.Note,
		})
	}

	return Summary{
		FilePath:     sum.FilePath,
		Pass:         sum.Pass,
		Warn:         sum.Warn,
		Fail:         sum.Fail,
		Errors:       sum.Error,
		AppliedFixes: sum.AppliedFixes,
		EarlyExit:    sum.EarlyExit,
		EarlyCheck:   sum.EarlyCheck,
		EarlyStatus:  string(sum.EarlyStatus),
		FinalPath:    sum.FinalPath,
		Outcomes:     outcomes,
	}
}

func isCriticalCheck(name string) bool {
	cu, ok := checks.Lookup(name)
	return ok && cu.FailFast()
}
