package checks

import (
	"fmt"
	"path/filepath"
	"strings"
)

const ensCSVName = "ensure-csv-extension"

type ensureCSV struct{}

func (ensureCSV) Name() string   { return ensCSVName }
func (ensureCSV) FailFast() bool { return true }
func (ensureCSV) Priority() int  { return 1 }

func (ensureCSV) Run(_data []byte, filePath string, _langs []string) Result {
	fp := strings.TrimSpace(filePath)
	ext := filepath.Ext(fp)

	if strings.EqualFold(ext, ".csv") {
		return Result{
			Name:    ensCSVName,
			Status:  Pass,
			Message: "File extension OK: .csv",
		}
	}

	if ext == "" {
		ext = "(none)"
	}

	return Result{
		Name:    ensCSVName,
		Status:  Fail,
		Message: fmt.Sprintf("Invalid file extension: %s (expected .csv)", ext),
	}
}

func init() { Register(ensureCSV{}) }
