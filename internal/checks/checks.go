package checks

import (
	"sort"
)

type Status string

const (
	Pass  Status = "PASS"
	Fail  Status = "FAIL"
	Error Status = "ERROR"
)

type Result struct {
	Name    string
	Status  Status
	Message string
}

type Check interface {
	Name() string
	Run(filePath string) Result
	// If true, stop the whole validation on non-PASS result for this check.
	FailFast() bool
	// Less number -> higher priority
	Priority() int
}

// Global registry of checks.
var All []Check

func Register(c Check) {
	All = append(All, c)
}

func Sorted() []Check {
	out := make([]Check, len(All))
	copy(out, All)
	sort.SliceStable(out, func(i, j int) bool {
		pi, pj := out[i].Priority(), out[j].Priority()
		if pi != pj {
			return pi < pj
		}

		return out[i].Name() < out[j].Name()
	})
	return out
}

func Reset() {
	All = nil
}
