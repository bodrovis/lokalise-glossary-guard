package guard

type ValidateRequest struct {
	Path string `json:"path,omitzero"`

	// Data is for Go callers: CLI, tests, server-side integrations.
	Data []byte `json:"-"`

	// Text is for JSON/WASM/browser callers.
	// It uses "data" in JSON because that's what the browser API naturally sends.
	Text string `json:"data,omitzero"`

	Langs []string `json:"langs,omitempty"`

	Fix           bool  `json:"fix"`
	RerunAfterFix *bool `json:"rerun_after_fix,omitzero"`
	HardFailOnErr bool  `json:"hard_fail_on_error"`
}

type ValidateStatus string

const (
	StatusPassed             ValidateStatus = "passed"
	StatusPassedWithWarnings ValidateStatus = "passed_with_warnings"
	StatusFailed             ValidateStatus = "failed"
)

type ValidateResponse struct {
	Path   string         `json:"path,omitzero"`
	Status ValidateStatus `json:"status"`

	Passed  bool `json:"passed"`
	Warned  bool `json:"warned"`
	Failed  bool `json:"failed"`
	Errored bool `json:"errored"`

	Error string `json:"error,omitzero"`

	Fixed     bool   `json:"fixed"`
	FixedText string `json:"fixed_text,omitzero"`

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
	EarlyCheck  string `json:"early_check,omitzero"`
	EarlyStatus string `json:"early_status,omitzero"`

	FinalPath string `json:"final_path,omitzero"`

	Outcomes []Outcome `json:"outcomes"`
}

type Outcome struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitzero"`

	Critical bool `json:"critical"`
	Changed  bool `json:"changed"`

	Note string `json:"note,omitzero"`
}
