// validator_test.go
package validator

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard/internal/checks"
)

/*** test doubles ***/

type mockCheck struct {
	name     string
	priority int
	failFast bool
	result   checks.Result

	// optional behavior controls
	panicOnRun bool
	expectLang string // if non-empty, will FAIL unless expectLang is present in langs
}

func (m mockCheck) Name() string   { return m.name }
func (m mockCheck) Priority() int  { return m.priority }
func (m mockCheck) FailFast() bool { return m.failFast }
func (m mockCheck) Run(_data []byte, _path string, langs []string) checks.Result {
	if m.panicOnRun {
		panic("kaboom")
	}
	if m.expectLang != "" {
		found := false
		for _, l := range langs {
			if l == m.expectLang {
				found = true
				break
			}
		}
		if !found {
			return checks.Failf(m.name, "missing expected lang %q", m.expectLang)
		}
	}
	if (m.result == checks.Result{}) {
		return checks.Passf(m.name, "ok")
	}
	// ensure Name is set for realism
	if m.result.Name == "" {
		mr := m.result
		mr.Name = m.name
		return mr
	}
	return m.result
}

/*** helpers ***/

func dataFixture() []byte {
	return []byte("term;description\na;b\n")
}

/*** tests ***/

func TestValidate_NoChecksRegistered(t *testing.T) {
	checks.Reset()

	sum, err := Validate(dataFixture(), "file.csv", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Pass != 0 || sum.Fail != 0 || sum.Error != 0 || len(sum.Results) != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestValidate_AllPass(t *testing.T) {
	checks.Reset()

	// critical pass
	checks.Register(mockCheck{
		name:     "crit-pass",
		priority: 10,
		failFast: true,
		result:   checks.Result{Status: checks.Pass, Message: "ok"},
	})
	// normal pass
	checks.Register(mockCheck{
		name:     "norm-pass",
		priority: 20,
		failFast: false,
		result:   checks.Result{Status: checks.Pass, Message: "ok"},
	})

	sum, err := Validate(dataFixture(), "x.csv", []string{"en"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Pass != 2 || sum.Fail != 0 || sum.Error != 0 || len(sum.Results) != 2 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
	// critical first, then normal (normal block may be sorted, but with single item it's trivial)
	if sum.Results[0].Name != "crit-pass" || sum.Results[1].Name != "norm-pass" {
		t.Fatalf("unexpected order: %+v", sum.Results)
	}
}

func TestValidate_FailFastStopsEarly(t *testing.T) {
	checks.Reset()

	// First critical fails and should stop early.
	checks.Register(mockCheck{
		name:     "crit-fail",
		priority: 1,
		failFast: true,
		result:   checks.Result{Status: checks.Fail, Message: "nope"},
	})
	// Would pass, but must not run due to early exit.
	checks.Register(mockCheck{
		name:     "norm-pass",
		priority: 100,
		failFast: false,
		result:   checks.Result{Status: checks.Pass, Message: "ok"},
	})

	sum, err := Validate(dataFixture(), "file.csv", nil)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
	if sum.EarlyExit != true || sum.EarlyCheck != "crit-fail" || sum.EarlyStatus != checks.Fail {
		t.Fatalf("expected early exit info set, got: %+v", sum)
	}
	if sum.Pass != 0 || sum.Fail != 1 || sum.Error != 0 {
		t.Fatalf("unexpected tallies: %+v", sum)
	}
	if len(sum.Results) != 1 || sum.Results[0].Name != "crit-fail" {
		t.Fatalf("unexpected results: %+v", sum.Results)
	}
}

func TestValidate_ErrorFailFastStopsEarly(t *testing.T) {
	checks.Reset()

	checks.Register(mockCheck{
		name:     "crit-error",
		priority: 1,
		failFast: true,
		result:   checks.Result{Status: checks.Error, Message: "boom"},
	})
	// normal shouldn't run
	checks.Register(mockCheck{
		name:     "norm-pass",
		priority: 2,
		failFast: false,
		result:   checks.Result{Status: checks.Pass},
	})

	sum, err := Validate(dataFixture(), "file.csv", nil)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
	if sum.Pass != 0 || sum.Fail != 0 || sum.Error != 1 || len(sum.Results) != 1 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
	if sum.Results[0].Name != "crit-error" || sum.Results[0].Status != checks.Error {
		t.Fatalf("unexpected first result: %+v", sum.Results[0])
	}
}

func TestValidate_NormalFailuresAreCollectedAndSorted(t *testing.T) {
	checks.Reset()

	// critical pass to allow normal checks to run
	checks.Register(mockCheck{
		name:     "crit-pass",
		priority: 1,
		failFast: true,
		result:   checks.Result{Status: checks.Pass},
	})

	// two normal checks inserted in reverse name order to verify deterministic sort
	checks.Register(mockCheck{
		name:     "zzz-normal-fail",
		priority: 100,
		failFast: false,
		result:   checks.Result{Status: checks.Fail, Message: "bad"},
	})
	checks.Register(mockCheck{
		name:     "aaa-normal-fail",
		priority: 100,
		failFast: false,
		result:   checks.Result{Status: checks.Fail, Message: "worse"},
	})

	sum, err := Validate(dataFixture(), "file.csv", nil)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}

	// Expect 1 critical + 2 normal
	if len(sum.Results) != 3 {
		t.Fatalf("unexpected results len: %d", len(sum.Results))
	}

	// normals should be sorted by name (then by status)
	normal := sum.Results[1:]
	got := []string{normal[0].Name, normal[1].Name}
	want := []string{"aaa-normal-fail", "zzz-normal-fail"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normal results not sorted deterministically\n got: %v\nwant: %v", got, want)
	}
}

func TestValidate_RecoveryFromPanicCountsAsError(t *testing.T) {
	checks.Reset()

	// critical pass
	checks.Register(mockCheck{
		name:     "crit-pass",
		priority: 1,
		failFast: true,
		result:   checks.Result{Status: checks.Pass},
	})

	// normal panics
	checks.Register(mockCheck{
		name:       "norm-panic",
		priority:   2,
		failFast:   false,
		panicOnRun: true,
	})

	sum, err := Validate(dataFixture(), "file.csv", nil)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed due to ERROR, got %v", err)
	}

	if sum.Error != 1 {
		t.Fatalf("expected 1 error, got %d", sum.Error)
	}

	var names []string
	for _, r := range sum.Results {
		names = append(names, r.Name)
	}
	sort.Strings(names)
	if names[0] != "crit-pass" || names[1] != "norm-panic" {
		t.Fatalf("unexpected result names: %v", names)
	}
}

func TestValidate_LangsArePassedThrough(t *testing.T) {
	checks.Reset()

	checks.Register(mockCheck{
		name:     "crit-pass",
		priority: 1,
		failFast: true,
		result:   checks.Result{Status: checks.Pass},
	})
	// normal requires a particular language to PASS
	checks.Register(mockCheck{
		name:       "norm-lang-sensitive",
		priority:   2,
		failFast:   false,
		expectLang: "de_DE",
	})

	// Missing expected lang -> should FAIL
	sum, err := Validate(dataFixture(), "file.csv", []string{"en"})
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed when lang missing; got %v", err)
	}
	if sum.Fail == 0 {
		t.Fatalf("expected a failure due to missing lang propagation")
	}

	// With expected lang -> should PASS
	checks.Reset()
	checks.Register(mockCheck{name: "crit-pass", priority: 1, failFast: true, result: checks.Result{Status: checks.Pass}})
	checks.Register(mockCheck{name: "norm-lang-sensitive", priority: 2, failFast: false, expectLang: "de_DE", result: checks.Result{Status: checks.Pass}})

	sum, err = Validate(dataFixture(), "file.csv", []string{"en", "de_DE"})
	if err != nil {
		t.Fatalf("unexpected error with correct langs: %v", err)
	}
	if sum.Fail != 0 || sum.Error != 0 {
		t.Fatalf("unexpected summary with langs: %+v", sum)
	}
}
