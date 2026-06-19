package validate

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestApplyGuardResponse_Passed(t *testing.T) {
	var oc fileOutcome

	summary := guard.Summary{
		FilePath: "ok.csv",
		Pass:     10,
	}

	applyGuardResponse(&oc, guard.ValidateResponse{
		Passed:  true,
		Summary: summary,
	})

	if oc.Passed != 1 {
		t.Fatalf("Passed = %d, want 1", oc.Passed)
	}

	if oc.Warned != 0 {
		t.Fatalf("Warned = %d, want 0", oc.Warned)
	}

	if oc.Failed != 0 {
		t.Fatalf("Failed = %d, want 0", oc.Failed)
	}

	if oc.Errored != 0 {
		t.Fatalf("Errored = %d, want 0", oc.Errored)
	}

	if oc.HadValFail {
		t.Fatal("HadValFail = true, want false")
	}

	if oc.Summary == nil {
		t.Fatal("Summary = nil, want summary")
	}

	if oc.Summary.FilePath != "ok.csv" {
		t.Fatalf("Summary.FilePath = %q, want %q", oc.Summary.FilePath, "ok.csv")
	}
}

func TestApplyGuardResponse_Warned(t *testing.T) {
	var oc fileOutcome

	applyGuardResponse(&oc, guard.ValidateResponse{
		Warned: true,
		Summary: guard.Summary{
			FilePath: "warn.csv",
			Warn:     2,
		},
	})

	if oc.Passed != 0 {
		t.Fatalf("Passed = %d, want 0", oc.Passed)
	}

	if oc.Warned != 1 {
		t.Fatalf("Warned = %d, want 1", oc.Warned)
	}

	if oc.Failed != 0 {
		t.Fatalf("Failed = %d, want 0", oc.Failed)
	}

	if oc.HadValFail {
		t.Fatal("HadValFail = true, want false")
	}

	if oc.Summary == nil || oc.Summary.Warn != 2 {
		t.Fatalf("Summary = %#v, want Warn=2", oc.Summary)
	}
}

func TestApplyGuardResponse_Failed(t *testing.T) {
	var oc fileOutcome

	applyGuardResponse(&oc, guard.ValidateResponse{
		Failed: true,
		Summary: guard.Summary{
			FilePath: "bad.csv",
			Fail:     1,
		},
	})

	if oc.Passed != 0 {
		t.Fatalf("Passed = %d, want 0", oc.Passed)
	}

	if oc.Warned != 0 {
		t.Fatalf("Warned = %d, want 0", oc.Warned)
	}

	if oc.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", oc.Failed)
	}

	if !oc.HadValFail {
		t.Fatal("HadValFail = false, want true")
	}

	if oc.Summary == nil || oc.Summary.Fail != 1 {
		t.Fatalf("Summary = %#v, want Fail=1", oc.Summary)
	}
}

func TestApplyGuardResponse_Errored(t *testing.T) {
	var oc fileOutcome

	applyGuardResponse(&oc, guard.ValidateResponse{
		Failed:  true,
		Errored: true,
		Summary: guard.Summary{
			FilePath: "errored.csv",
			Errors:   1,
		},
	})

	if oc.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", oc.Failed)
	}

	if !oc.HadValFail {
		t.Fatal("HadValFail = false, want true")
	}

	// This expects applyGuardResponse to map resp.Errored into oc.Errored.
	// Add:
	//   if resp.Errored { oc.Errored = 1 }
	if oc.Errored != 1 {
		t.Fatalf("Errored = %d, want 1", oc.Errored)
	}
}

func TestFileOpErrorOutcome(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	var b strings.Builder

	oc := fileOpErrorOutcome(
		7,
		"missing.csv",
		&b,
		errors.New("open failed"),
		"---",
	)

	if oc.Idx != 7 {
		t.Fatalf("Idx = %d, want 7", oc.Idx)
	}

	if oc.Path != "missing.csv" {
		t.Fatalf("Path = %q, want %q", oc.Path, "missing.csv")
	}

	if oc.Errored != 1 {
		t.Fatalf("Errored = %d, want 1", oc.Errored)
	}

	if !oc.HadOpErr {
		t.Fatal("HadOpErr = false, want true")
	}

	wantOutput := "ERROR: open failed\n---\n"
	if oc.Output != wantOutput {
		t.Fatalf("Output = %q, want %q", oc.Output, wantOutput)
	}

	if b.String() != wantOutput {
		t.Fatalf("builder output = %q, want %q", b.String(), wantOutput)
	}
}

func TestAggregateOutcomes_Empty(t *testing.T) {
	got := aggregateOutcomes(nil)

	if got != (aggregateResult{}) {
		t.Fatalf("aggregateOutcomes(nil) = %#v, want zero result", got)
	}
}

func TestAggregateOutcomes_CountsAndFlags(t *testing.T) {
	outcomes := []fileOutcome{
		{
			Path:   "ok.csv",
			Passed: 1,
			Summary: &guard.Summary{
				Warn: 2,
			},
		},
		{
			Path:       "bad.csv",
			Failed:     1,
			HadValFail: true,
			Summary: &guard.Summary{
				Warn: 3,
			},
		},
		{
			Path:     "missing.csv",
			Errored:  1,
			HadOpErr: true,
		},
	}

	got := aggregateOutcomes(outcomes)

	want := aggregateResult{
		HadOpErr:   true,
		HadValFail: true,
		Passed:     1,
		Warns:      5,
		Failed:     1,
		Errored:    1,
	}

	if got != want {
		t.Fatalf("aggregateOutcomes = %#v, want %#v", got, want)
	}
}

func TestAggregateReturnCode_NilWhenNoErrors(t *testing.T) {
	err := aggregateReturnCode([]fileOutcome{
		{Path: "ok.csv", Passed: 1},
		{Path: "warn.csv", Warned: 1},
	})

	if err != nil {
		t.Fatalf("aggregateReturnCode returned error: %v", err)
	}
}

func TestAggregateReturnCode_ValidationFailure(t *testing.T) {
	err := aggregateReturnCode([]fileOutcome{
		{
			Path:       "bad.csv",
			Failed:     1,
			HadValFail: true,
		},
	})

	if err == nil {
		t.Fatal("error = nil, want validation failed")
	}

	if err.Error() != "validation failed" {
		t.Fatalf("error = %q, want %q", err.Error(), "validation failed")
	}
}

func TestAggregateReturnCode_OperationalError(t *testing.T) {
	err := aggregateReturnCode([]fileOutcome{
		{
			Path:     "missing.csv",
			Errored:  1,
			HadOpErr: true,
		},
	})

	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	want := "one or more files could not be validated due to an error"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestAggregateReturnCode_OperationalErrorTakesPrecedence(t *testing.T) {
	err := aggregateReturnCode([]fileOutcome{
		{
			Path:       "missing.csv",
			Failed:     1,
			Errored:    1,
			HadOpErr:   true,
			HadValFail: true,
		},
	})

	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	want := "one or more files could not be validated due to an error"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestPrintAndAggregate_SingleFile(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:   "ok.csv",
			Output: "file output\n",
			Passed: 1,
			Summary: &guard.Summary{
				Warn: 2,
			},
		},
	}

	var (
		hadOpErr     bool
		hadValFail   bool
		filesPassed  int
		filesFailed  int
		filesErrored int
	)

	stdout := captureStdoutForOutcomeTest(t, func() {
		hadOpErr, hadValFail, filesPassed, filesFailed, filesErrored =
			printAndAggregate(outcomes, 1, time.Now())
	})

	if hadOpErr {
		t.Fatal("hadOpErr = true, want false")
	}

	if hadValFail {
		t.Fatal("hadValFail = true, want false")
	}

	if filesPassed != 1 {
		t.Fatalf("filesPassed = %d, want 1", filesPassed)
	}

	if filesFailed != 0 {
		t.Fatalf("filesFailed = %d, want 0", filesFailed)
	}

	if filesErrored != 0 {
		t.Fatalf("filesErrored = %d, want 0", filesErrored)
	}

	if !strings.Contains(stdout, "file output\n") {
		t.Fatalf("stdout = %q, want file output", stdout)
	}

	if strings.Contains(stdout, "Overall:") {
		t.Fatalf("stdout = %q, want no overall line for single file", stdout)
	}

	if !strings.Contains(stdout, "Total time:") {
		t.Fatalf("stdout = %q, want total time", stdout)
	}
}

func TestPrintAndAggregate_MultipleFiles(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:   "ok.csv",
			Output: "ok output\n",
			Passed: 1,
			Summary: &guard.Summary{
				Warn: 2,
			},
		},
		{
			Path:       "bad.csv",
			Output:     "bad output\n",
			Failed:     1,
			HadValFail: true,
			Summary: &guard.Summary{
				Warn: 1,
			},
		},
		{
			Path:     "missing.csv",
			Output:   "missing output\n",
			Errored:  1,
			HadOpErr: true,
		},
	}

	var (
		hadOpErr     bool
		hadValFail   bool
		filesPassed  int
		filesFailed  int
		filesErrored int
	)

	stdout := captureStdoutForOutcomeTest(t, func() {
		hadOpErr, hadValFail, filesPassed, filesFailed, filesErrored =
			printAndAggregate(outcomes, 3, time.Now())
	})

	if !hadOpErr {
		t.Fatal("hadOpErr = false, want true")
	}

	if !hadValFail {
		t.Fatal("hadValFail = false, want true")
	}

	if filesPassed != 1 {
		t.Fatalf("filesPassed = %d, want 1", filesPassed)
	}

	if filesFailed != 1 {
		t.Fatalf("filesFailed = %d, want 1", filesFailed)
	}

	if filesErrored != 1 {
		t.Fatalf("filesErrored = %d, want 1", filesErrored)
	}

	for _, want := range []string{
		"ok output\n",
		"bad output\n",
		"missing output\n",
		"Overall: 1 passed, 3 warning(s), 1 failed, 1 error(s)",
		"Total time:",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func captureStdoutForOutcomeTest(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("pipe writer close failed: %v", err)
	}

	os.Stdout = oldStdout

	out := <-done

	if err := r.Close(); err != nil {
		t.Fatalf("pipe reader close failed: %v", err)
	}

	return out
}
