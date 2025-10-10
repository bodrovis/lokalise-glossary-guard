package checks

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ensureCSV struct{}

func (ensureCSV) Name() string   { return "ensure-csv-extension" }
func (ensureCSV) FailFast() bool { return true }
func (ensureCSV) Priority() int  { return 1 }
func (ensureCSV) Run(filePath string) Result {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".csv" {
		return Result{
			Name:    "ensure-csv-extension",
			Status:  Pass,
			Message: "File extension OK: .csv",
		}
	}
	if ext == "" {
		ext = "(none)"
	}
	return Result{
		Name:    "ensure-csv-extension",
		Status:  Fail,
		Message: fmt.Sprintf("Invalid file extension: %s (expected .csv)", ext),
	}
}

func init() { Register(ensureCSV{}) }
