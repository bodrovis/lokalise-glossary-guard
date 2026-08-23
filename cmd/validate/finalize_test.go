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

type failingWriter struct{}

func (f failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestFinalizeJSON_EncodesOutcomes(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Idx:     42,
			Path:    "ok.csv",
			Output:  "human output must not be serialized",
			Passed:  1,
			Summary: &guard.Summary{Warn: 2},
		},
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	err := finalizeJSON(&out, &errOut, outcomes)
	if err != nil {
		t.Fatalf("finalizeJSON returned error: %v", err)
	}

	got := out.String()

	for _, want := range []string{
		"[\n",
		`"path": "ok.csv"`,
		`"passed": 1`,
		`"warned": 0`,
		`"failed": 0`,
		`"errored": 0`,
		`"summary": {`,
		`"warn": 2`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("json output = %q, want it to contain %q", got, want)
		}
	}

	for _, unwanted := range []string{
		`"Idx"`,
		`"idx"`,
		`"Output"`,
		`"output"`,
		"human output must not be serialized",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("json output = %q, want it not to contain %q", got, unwanted)
		}
	}

	if errOut.String() != "" {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestFinalizeJSON_ReturnsValidationFailureAfterEncoding(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:       "bad.csv",
			Failed:     1,
			HadValFail: true,
			Summary:    &guard.Summary{Fail: 1},
		},
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	err := finalizeJSON(&out, &errOut, outcomes)
	if err == nil {
		t.Fatal("error = nil, want validation failed")
	}

	if err.Error() != "validation failed" {
		t.Fatalf("error = %q, want %q", err.Error(), "validation failed")
	}

	if !strings.Contains(out.String(), `"path": "bad.csv"`) {
		t.Fatalf("json output = %q, want encoded outcomes", out.String())
	}

	if errOut.String() != "" {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestFinalizeJSON_ReturnsOperationalErrorAfterEncoding(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:     "missing.csv",
			Errored:  1,
			HadOpErr: true,
		},
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	err := finalizeJSON(&out, &errOut, outcomes)
	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	want := "one or more files could not be validated due to an error"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}

	if !strings.Contains(out.String(), `"path": "missing.csv"`) {
		t.Fatalf("json output = %q, want encoded outcomes", out.String())
	}

	if errOut.String() != "" {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestFinalizeJSON_OperationalErrorTakesPrecedence(t *testing.T) {
	outcomes := []fileOutcome{
		{
			Path:       "missing.csv",
			Errored:    1,
			Failed:     1,
			HadOpErr:   true,
			HadValFail: true,
		},
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	err := finalizeJSON(&out, &errOut, outcomes)
	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	want := "one or more files could not be validated due to an error"
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestFinalizeJSON_ReturnsEncodeError(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	var errOut bytes.Buffer

	err := finalizeJSON(failingWriter{}, &errOut, []fileOutcome{
		{Path: "ok.csv", Passed: 1},
	})

	if err == nil {
		t.Fatal("error = nil, want encode error")
	}

	if !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("error = %q, want write failed", err.Error())
	}

	want := "failed to encode json: jsontext: write error"
	if !strings.Contains(errOut.String(), want) {
		t.Fatalf("stderr = %q, want it to contain %q", errOut.String(), want)
	}
}

func TestFinalizeText_ReturnsNilWhenNoFailures(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:   "ok.csv",
			Passed: 1,
			Output: "file output\n",
		},
	}

	stdout := captureStdout(t, func() {
		err := finalizeText(outcomes, 1, time.Now())
		if err != nil {
			t.Fatalf("finalizeText returned error: %v", err)
		}
	})

	if !strings.Contains(stdout, "file output\n") {
		t.Fatalf("stdout = %q, want file output", stdout)
	}

	if !strings.Contains(stdout, "Total time:") {
		t.Fatalf("stdout = %q, want total time", stdout)
	}

	if strings.Contains(stdout, "Overall:") {
		t.Fatalf("stdout = %q, want no overall line for single file", stdout)
	}
}

func TestFinalizeText_ReturnsValidationFailed(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:       "bad.csv",
			Failed:     1,
			HadValFail: true,
			Output:     "bad file output\n",
		},
	}

	var gotErr error
	stdout := captureStdout(t, func() {
		gotErr = finalizeText(outcomes, 1, time.Now())
	})

	if gotErr == nil {
		t.Fatal("error = nil, want validation failed")
	}

	if gotErr.Error() != "validation failed" {
		t.Fatalf("error = %q, want %q", gotErr.Error(), "validation failed")
	}

	if !strings.Contains(stdout, "bad file output\n") {
		t.Fatalf("stdout = %q, want file output", stdout)
	}
}

func TestFinalizeText_ReturnsOperationalError(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:     "missing.csv",
			Errored:  1,
			HadOpErr: true,
			Output:   "read error output\n",
		},
	}

	var gotErr error
	stdout := captureStdout(t, func() {
		gotErr = finalizeText(outcomes, 1, time.Now())
	})

	if gotErr == nil {
		t.Fatal("error = nil, want operational error")
	}

	want := "one or more files could not be validated due to an error"
	if gotErr.Error() != want {
		t.Fatalf("error = %q, want %q", gotErr.Error(), want)
	}

	if !strings.Contains(stdout, "read error output\n") {
		t.Fatalf("stdout = %q, want file output", stdout)
	}
}

func TestFinalizeText_OperationalErrorTakesPrecedence(t *testing.T) {
	outcomes := []fileOutcome{
		{
			Path:       "missing.csv",
			Failed:     1,
			Errored:    1,
			HadOpErr:   true,
			HadValFail: true,
		},
	}

	var gotErr error
	_ = captureStdout(t, func() {
		gotErr = finalizeText(outcomes, 1, time.Now())
	})

	if gotErr == nil {
		t.Fatal("error = nil, want operational error")
	}

	want := "one or more files could not be validated due to an error"
	if gotErr.Error() != want {
		t.Fatalf("error = %q, want %q", gotErr.Error(), want)
	}
}

func TestFinalizeText_PrintsOverallForMultipleFiles(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	outcomes := []fileOutcome{
		{
			Path:   "ok.csv",
			Passed: 1,
			Summary: &guard.Summary{
				Warn: 2,
			},
		},
		{
			Path:   "bad.csv",
			Failed: 1,
			Summary: &guard.Summary{
				Warn: 1,
			},
			HadValFail: true,
		},
		{
			Path:     "missing.csv",
			Errored:  1,
			HadOpErr: true,
		},
	}

	var gotErr error
	stdout := captureStdout(t, func() {
		gotErr = finalizeText(outcomes, 3, time.Now())
	})

	if gotErr == nil {
		t.Fatal("error = nil, want operational error")
	}

	wantOverall := "Overall: 1 passed, 3 warning(s), 1 failed, 1 error(s)"
	if !strings.Contains(stdout, wantOverall) {
		t.Fatalf("stdout = %q, want overall line %q", stdout, wantOverall)
	}
}

func captureStdout(t *testing.T, fn func()) string {
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
