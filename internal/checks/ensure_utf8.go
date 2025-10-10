package checks

import (
	"fmt"
	"os"
	"unicode/utf8"
)

type ensureUTF8 struct{}

func (ensureUTF8) Name() string   { return "ensure-utf8-encoding" }
func (ensureUTF8) FailFast() bool { return true }
func (ensureUTF8) Priority() int  { return 2 }

func (ensureUTF8) Run(filePath string) Result {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Result{
			Name:    "ensure-utf8-encoding",
			Status:  Error,
			Message: fmt.Sprintf("cannot read file: %v", err),
		}
	}

	if utf8.Valid(data) {
		return Result{
			Name:    "ensure-utf8-encoding",
			Status:  Pass,
			Message: "File encoding is valid UTF-8",
		}
	}

	return Result{
		Name:    "ensure-utf8-encoding",
		Status:  Fail,
		Message: "File encoding is not valid UTF-8",
	}
}

func init() { Register(ensureUTF8{}) }
