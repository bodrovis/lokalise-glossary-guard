package validator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/bodrovis/lokalise-glossary-guard/internal/checks"
)

var ErrValidationFailed = errors.New("validation failed")

type Summary struct {
	FilePath    string
	Pass        int
	Fail        int
	Error       int
	Results     []checks.Result
	EarlyExit   bool
	EarlyCheck  string
	EarlyStatus checks.Status
}

func safeRun(c checks.Check, filePath string) (res checks.Result) {
	defer func() {
		if r := recover(); r != nil {
			res.Status = checks.Error
			res.Message = fmt.Sprintf("check panicked: %v", r)
		}
	}()
	return c.Run(filePath)
}

func Validate(filePath string) (Summary, error) {
	if err := ValidateFilePath(filePath); err != nil {
		return Summary{}, err
	}

	ordered := checks.Sorted()
	if len(ordered) == 0 {
		return Summary{FilePath: filePath, Results: nil}, nil
	}

	sum := Summary{
		FilePath: filePath,
		Results:  make([]checks.Result, 0, len(ordered)),
	}

	for _, c := range ordered {
		res := safeRun(c, filePath)
		sum.Results = append(sum.Results, res)

		switch res.Status {
		case checks.Pass:
			sum.Pass++
		case checks.Fail:
			sum.Fail++
			if c.FailFast() {
				sum.EarlyExit = true
				sum.EarlyCheck = c.Name()
				sum.EarlyStatus = res.Status
				return sum, ErrValidationFailed
			}
		case checks.Error:
			sum.Error++
			if c.FailFast() {
				sum.EarlyExit = true
				sum.EarlyCheck = c.Name()
				sum.EarlyStatus = res.Status
				return sum, ErrValidationFailed
			}
		default:
			sum.Error++
			if c.FailFast() {
				sum.EarlyExit = true
				sum.EarlyCheck = c.Name()
				sum.EarlyStatus = res.Status
				return sum, ErrValidationFailed
			}
		}
	}

	if sum.Fail > 0 || sum.Error > 0 {
		return sum, ErrValidationFailed
	}
	return sum, nil
}

func ValidateFilePath(fp string) error {
	if fp == "" {
		return fmt.Errorf("file path is required")
	}
	info, err := os.Stat(fp)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return fmt.Errorf("file not found: %s", fp)
		case errors.Is(err, fs.ErrPermission):
			return fmt.Errorf("permission denied: %s", fp)
		default:
			return fmt.Errorf("file not accessible: %w", err)
		}
	}
	if info.IsDir() {
		return fmt.Errorf("path points to a directory: %s", fp)
	}
	f, err := os.Open(fp)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("permission denied: %s", fp)
		}
		return fmt.Errorf("file is not readable: %w", err)
	}
	_ = f.Close()
	return nil
}
