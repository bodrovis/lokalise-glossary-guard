package validator

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard/internal/checks"
)

/*** test doubles ***/

type mockCheck struct {
	name     string
	priority int
	failFast bool
	result   checks.Result
}

func (m mockCheck) Name() string               { return m.name }
func (m mockCheck) Priority() int              { return m.priority }
func (m mockCheck) FailFast() bool             { return m.failFast }
func (m mockCheck) Run(_ string) checks.Result { return m.result }

/*** helpers ***/

func tmpFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "file.csv")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

/*** tests ***/

func TestValidateFilePath(t *testing.T) {
	// empty path
	if err := ValidateFilePath(""); err == nil {
		t.Fatal("expected error for empty path")
	}

	// non-existent
	if err := ValidateFilePath(filepath.Join(t.TempDir(), "nope.csv")); err == nil {
		t.Fatal("expected error for missing file")
	}

	// directory
	dir := t.TempDir()
	if err := ValidateFilePath(dir); err == nil {
		t.Fatal("expected error for directory path")
	}

	// readable file ok
	p := tmpFile(t, "ok")
	if err := ValidateFilePath(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_NoChecksRegistered(t *testing.T) {
	checks.Reset()
	p := tmpFile(t, "doesn't matter")
	sum, err := Validate(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Pass != 0 || sum.Fail != 0 || sum.Error != 0 || len(sum.Results) != 0 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestValidate_AllPass(t *testing.T) {
	checks.Reset()

	checks.Register(mockCheck{
		name:     "pass-1",
		priority: 10,
		failFast: true,
		result:   checks.Result{Name: "pass-1", Status: checks.Pass, Message: "ok"},
	})
	checks.Register(mockCheck{
		name:     "pass-2",
		priority: 20,
		failFast: false,
		result:   checks.Result{Name: "pass-2", Status: checks.Pass, Message: "ok"},
	})

	p := tmpFile(t, "content")
	sum, err := Validate(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Pass != 2 || sum.Fail != 0 || sum.Error != 0 || len(sum.Results) != 2 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestValidate_FailFastStopsEarly(t *testing.T) {
	checks.Reset()

	// First check fails and is fail-fast.
	checks.Register(mockCheck{
		name:     "fail-fast",
		priority: 1,
		failFast: true,
		result:   checks.Result{Name: "fail-fast", Status: checks.Fail, Message: "nope"},
	})
	// Would pass, but should never run.
	checks.Register(mockCheck{
		name:     "pass-late",
		priority: 2,
		failFast: false,
		result:   checks.Result{Name: "pass-late", Status: checks.Pass, Message: "ok"},
	})

	p := tmpFile(t, "x")
	sum, err := Validate(p)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
	if sum.Pass != 0 || sum.Fail != 1 || sum.Error != 0 || len(sum.Results) != 1 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
	if sum.Results[0].Name != "fail-fast" {
		t.Fatalf("unexpected first result: %+v", sum.Results[0])
	}
}

func TestValidate_NonFailFastFailureContinues(t *testing.T) {
	checks.Reset()

	checks.Register(mockCheck{
		name:     "fail-nonfast",
		priority: 1,
		failFast: false,
		result:   checks.Result{Name: "fail-nonfast", Status: checks.Fail, Message: "bad"},
	})
	checks.Register(mockCheck{
		name:     "pass-after",
		priority: 2,
		failFast: true,
		result:   checks.Result{Name: "pass-after", Status: checks.Pass, Message: "ok"},
	})

	p := tmpFile(t, "x")
	sum, err := Validate(p)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
	if sum.Pass != 1 || sum.Fail != 1 || sum.Error != 0 || len(sum.Results) != 2 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}

func TestValidate_ErrorFailFastStopsEarly(t *testing.T) {
	checks.Reset()

	checks.Register(mockCheck{
		name:     "error-fast",
		priority: 1,
		failFast: true,
		result:   checks.Result{Name: "error-fast", Status: checks.Error, Message: "boom"},
	})
	checks.Register(mockCheck{
		name:     "pass-late",
		priority: 2,
		failFast: false,
		result:   checks.Result{Name: "pass-late", Status: checks.Pass, Message: "ok"},
	})

	p := tmpFile(t, "x")
	sum, err := Validate(p)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
	if sum.Pass != 0 || sum.Fail != 0 || sum.Error != 1 || len(sum.Results) != 1 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
	if sum.Results[0].Status != checks.Error {
		t.Fatalf("expected first to be ERROR, got %+v", sum.Results[0])
	}
}
