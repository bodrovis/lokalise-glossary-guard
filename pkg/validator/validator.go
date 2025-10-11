package validator

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sort"
	"sync"

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

func Validate(data []byte, filePath string, langs []string) (Summary, error) {
	criticalChecks, normalChecks := checks.Split()

	total := len(criticalChecks) + len(normalChecks)
	if total == 0 {
		return Summary{FilePath: filePath, Results: nil}, nil
	}

	sum := Summary{
		FilePath: filePath,
		Results:  make([]checks.Result, 0, total),
	}

	for _, c := range criticalChecks {
		r := safeRun(c, data, filePath, langs)

		sum.Results = append(sum.Results, r)

		tally(&sum, r)

		if r.Status != checks.Pass {
			sum.EarlyExit = true
			sum.EarlyCheck = c.Name()
			sum.EarlyStatus = r.Status
			return sum, ErrValidationFailed
		}
	}

	if len(normalChecks) > 0 {
		resCh := make(chan checks.Result, len(normalChecks))
		var wg sync.WaitGroup

		for _, c := range normalChecks {
			wg.Add(1)
			go func(c checks.Check) {
				defer wg.Done()
				resCh <- safeRun(c, data, filePath, langs)
			}(c)
		}

		wg.Wait()
		close(resCh)

		for r := range resCh {
			sum.Results = append(sum.Results, r)
			tally(&sum, r)
		}

		normStart := len(criticalChecks)
		normSlice := sum.Results[normStart:]
		sort.SliceStable(normSlice, func(i, j int) bool {
			ni, nj := normSlice[i], normSlice[j]
			if ni.Name != nj.Name {
				return ni.Name < nj.Name
			}

			return ni.Status < nj.Status
		})
	}

	if sum.Fail > 0 || sum.Error > 0 {
		return sum, ErrValidationFailed
	}
	return sum, nil
}

func safeRun(c checks.Check, data []byte, path string, langs []string) (out checks.Result) {
	defer func() {
		if rec := recover(); rec != nil {
			out = checks.Result{
				Name:    c.Name(),
				Status:  checks.Error,
				Message: fmt.Sprintf("panic: %v\n%s", rec, debug.Stack()),
			}
		}
	}()
	r := c.Run(data, path, langs)
	if r.Name == "" {
		r.Name = c.Name()
	}
	return r
}

func tally(sum *Summary, r checks.Result) {
	switch r.Status {
	case checks.Pass:
		sum.Pass++
	case checks.Fail:
		sum.Fail++
	case checks.Error:
		sum.Error++
	default:
		sum.Error++
	}
}
