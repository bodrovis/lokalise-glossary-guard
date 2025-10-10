package checks

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

type ensureNonEmptyTerm struct{}

func (ensureNonEmptyTerm) Name() string   { return "ensure-non-empty-term" }
func (ensureNonEmptyTerm) FailFast() bool { return false }
func (ensureNonEmptyTerm) Priority() int  { return 4 }

func (ensureNonEmptyTerm) Run(filePath string) Result {
	f, err := os.Open(filePath)
	if err != nil {
		return Result{Name: "ensure-non-empty-term", Status: Error, Message: fmt.Sprintf("cannot open file: %v", err)}
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close %s: %v\n", filePath, cerr)
		}
	}()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	// читаем хедер
	header, err := r.Read()
	if err != nil {
		return Result{Name: "ensure-non-empty-term", Status: Error, Message: fmt.Sprintf("cannot read header: %v", err)}
	}

	termIdx := -1
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), "term") {
			termIdx = i
			break
		}
	}
	if termIdx == -1 {
		return Result{Name: "ensure-non-empty-term", Status: Error, Message: "header does not contain 'term' column"}
	}

	line := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			return Result{Name: "ensure-non-empty-term", Status: Error, Message: fmt.Sprintf("csv parse error: %v", err)}
		}
		if len(rec) <= termIdx {
			continue
		}
		term := strings.TrimSpace(rec[termIdx])
		if term == "" {
			return Result{Name: "ensure-non-empty-term", Status: Fail, Message: fmt.Sprintf("term value is required (blank found at line %d)", line)}
		}
	}

	return Result{Name: "ensure-non-empty-term", Status: Pass, Message: "All term values are present"}
}

func init() { Register(ensureNonEmptyTerm{}) }
