package checks

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

type ensureNonEmptyTerm struct{}

const ensNonEmptyTermsName = "ensure-non-empty-term"

func (ensureNonEmptyTerm) Name() string   { return ensNonEmptyTermsName }
func (ensureNonEmptyTerm) FailFast() bool { return false }
func (ensureNonEmptyTerm) Priority() int  { return 100 }

func (ensureNonEmptyTerm) Run(data []byte, _filePath string, _langs []string) Result {
	if len(data) == 0 {
		return Result{Name: ensNonEmptyTermsName, Status: Fail, Message: "Empty file: cannot validate terms"}
	}

	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return Result{Name: ensNonEmptyTermsName, Status: Fail, Message: "File has no header or data"}
		}
		return Result{Name: ensNonEmptyTermsName, Status: Error, Message: fmt.Sprintf("cannot read header: %v", err)}
	}

	termIdx := -1
	for i, h := range header {
		hh := strings.ToLower(strings.TrimSpace(h))
		if i == 0 {
			hh = strings.TrimPrefix(hh, "\uFEFF") // remove BOM
		}
		if hh == "term" {
			termIdx = i
			break
		}
	}

	if termIdx == -1 {
		return Result{Name: ensNonEmptyTermsName, Status: Error, Message: "Header does not contain 'term' column"}
	}

	lineNum := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}

		lineNum++
		if err != nil {
			return Result{Name: ensNonEmptyTermsName, Status: Error, Message: fmt.Sprintf("CSV parse error at line %d: %v", lineNum, err)}
		}
		if len(rec) <= termIdx {
			continue
		}

		if strings.TrimSpace(rec[termIdx]) == "" {
			return Result{Name: ensNonEmptyTermsName, Status: Fail, Message: fmt.Sprintf("Term value is required (blank found at line %d)", lineNum)}
		}
	}

	return Result{Name: ensNonEmptyTermsName, Status: Pass, Message: "All 'term' values are non-empty"}
}

func init() { Register(ensureNonEmptyTerm{}) }
