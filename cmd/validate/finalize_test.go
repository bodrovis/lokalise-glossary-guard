package validate

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func testFinalizer(jsonOut bool) (
	finalizer,
	*bytes.Buffer,
	*bytes.Buffer,
) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)

	return finalizer{
		out:     out,
		errOut:  errOut,
		colors:  newColorizer(true),
		jsonOut: jsonOut,
	}, out, errOut
}

func TestFinalizerJSON_EncodesOutcomes(t *testing.T) {
	t.Parallel()

	outcomes := []fileOutcome{
		{
			Idx:     42,
			Path:    "ok.csv",
			Output:  "human output must not be serialized",
			Passed:  1,
			Summary: &guard.Summary{Warn: 2},
		},
	}

	f, out, errOut := testFinalizer(true)

	if err := f.finalizeJSON(outcomes); err != nil {
		t.Fatalf("finalizeJSON() error = %v", err)
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
			t.Fatalf(
				"json output = %q, want it to contain %q",
				got,
				want,
			)
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
			t.Fatalf(
				"json output = %q, want it not to contain %q",
				got,
				unwanted,
			)
		}
	}

	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestFinalizerJSON_ReturnsValidationFailureAfterEncoding(t *testing.T) {
	t.Parallel()

	f, out, errOut := testFinalizer(true)

	err := f.finalizeJSON([]fileOutcome{
		{
			Path:       "bad.csv",
			Failed:     1,
			HadValFail: true,
			Summary:    &guard.Summary{Fail: 1},
		},
	})

	if err == nil {
		t.Fatal("error = nil, want validation failed")
	}

	if err.Error() != "validation failed" {
		t.Fatalf(
			"error = %q, want %q",
			err.Error(),
			"validation failed",
		)
	}

	if !strings.Contains(out.String(), `"path": "bad.csv"`) {
		t.Fatalf(
			"json output = %q, want encoded outcomes",
			out.String(),
		)
	}

	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestFinalizerJSON_ReturnsOperationalErrorAfterEncoding(t *testing.T) {
	t.Parallel()

	f, out, errOut := testFinalizer(true)

	err := f.finalizeJSON([]fileOutcome{
		{
			Path:     "missing.csv",
			Errored:  1,
			HadOpErr: true,
		},
	})

	want := "one or more files could not be validated due to an error"

	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}

	if !strings.Contains(out.String(), `"path": "missing.csv"`) {
		t.Fatalf(
			"json output = %q, want encoded outcomes",
			out.String(),
		)
	}

	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestFinalizerJSON_OperationalErrorTakesPrecedence(t *testing.T) {
	t.Parallel()

	f, _, _ := testFinalizer(true)

	err := f.finalizeJSON([]fileOutcome{
		{
			Path:       "missing.csv",
			Failed:     1,
			Errored:    1,
			HadOpErr:   true,
			HadValFail: true,
		},
	})

	want := "one or more files could not be validated due to an error"

	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestFinalizerJSON_ReturnsEncodeError(t *testing.T) {
	t.Parallel()

	var errOut bytes.Buffer

	f := finalizer{
		out:     failingWriter{},
		errOut:  &errOut,
		colors:  newColorizer(true),
		jsonOut: true,
	}

	err := f.finalizeJSON([]fileOutcome{
		{Path: "ok.csv", Passed: 1},
	})

	if err == nil {
		t.Fatal("error = nil, want encode error")
	}

	if !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("error = %q, want write failed", err.Error())
	}

	if !strings.Contains(
		errOut.String(),
		"failed to encode json:",
	) {
		t.Fatalf(
			"stderr = %q, want encode error",
			errOut.String(),
		)
	}
}

func TestFinalizerText_ReturnsNilWhenNoFailures(t *testing.T) {
	t.Parallel()

	f, out, _ := testFinalizer(false)

	err := f.finalizeText(
		[]fileOutcome{
			{
				Path:   "ok.csv",
				Passed: 1,
				Output: "file output\n",
			},
		},
		1,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("finalizeText() error = %v", err)
	}

	got := out.String()

	if !strings.Contains(got, "file output\n") {
		t.Fatalf("output = %q, want file output", got)
	}

	if !strings.Contains(got, "Total time:") {
		t.Fatalf("output = %q, want total time", got)
	}

	if strings.Contains(got, "Overall:") {
		t.Fatalf(
			"output = %q, want no overall line for single file",
			got,
		)
	}
}

func TestFinalizerText_ReturnsValidationFailed(t *testing.T) {
	t.Parallel()

	f, out, _ := testFinalizer(false)

	err := f.finalizeText(
		[]fileOutcome{
			{
				Path:       "bad.csv",
				Failed:     1,
				HadValFail: true,
				Output:     "bad file output\n",
			},
		},
		1,
		time.Now(),
	)

	if err == nil {
		t.Fatal("error = nil, want validation failed")
	}

	if err.Error() != "validation failed" {
		t.Fatalf(
			"error = %q, want %q",
			err.Error(),
			"validation failed",
		)
	}

	if !strings.Contains(out.String(), "bad file output\n") {
		t.Fatalf(
			"output = %q, want file output",
			out.String(),
		)
	}
}

func TestFinalizerText_ReturnsOperationalError(t *testing.T) {
	t.Parallel()

	f, out, _ := testFinalizer(false)

	err := f.finalizeText(
		[]fileOutcome{
			{
				Path:     "missing.csv",
				Errored:  1,
				HadOpErr: true,
				Output:   "read error output\n",
			},
		},
		1,
		time.Now(),
	)

	want := "one or more files could not be validated due to an error"

	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}

	if !strings.Contains(out.String(), "read error output\n") {
		t.Fatalf(
			"output = %q, want file output",
			out.String(),
		)
	}
}

func TestFinalizerText_OperationalErrorTakesPrecedence(t *testing.T) {
	t.Parallel()

	f, _, _ := testFinalizer(false)

	err := f.finalizeText(
		[]fileOutcome{
			{
				Path:       "missing.csv",
				Failed:     1,
				Errored:    1,
				HadOpErr:   true,
				HadValFail: true,
			},
		},
		1,
		time.Now(),
	)

	want := "one or more files could not be validated due to an error"

	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestFinalizerText_PrintsOverallForMultipleFiles(t *testing.T) {
	t.Parallel()

	f, out, _ := testFinalizer(false)

	err := f.finalizeText(
		[]fileOutcome{
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
		},
		3,
		time.Now(),
	)

	if err == nil {
		t.Fatal("error = nil, want operational error")
	}

	want := "Overall: 1 passed, 3 warning(s), 1 failed, 1 error(s)"

	if !strings.Contains(out.String(), want) {
		t.Fatalf(
			"output = %q, want overall line %q",
			out.String(),
			want,
		)
	}
}

func TestFinalizer_Finalize(t *testing.T) {
	t.Parallel()

	t.Run("text", func(t *testing.T) {
		t.Parallel()

		f, out, _ := testFinalizer(false)

		if err := f.finalize(nil, 0, time.Now()); err != nil {
			t.Fatalf("finalize() error = %v", err)
		}

		if !strings.Contains(out.String(), "Total time:") {
			t.Fatalf(
				"output = %q, want total time",
				out.String(),
			)
		}
	})

	t.Run("json", func(t *testing.T) {
		t.Parallel()

		f, out, _ := testFinalizer(true)

		if err := f.finalize([]fileOutcome{}, 0, time.Now()); err != nil {
			t.Fatalf("finalize() error = %v", err)
		}

		if strings.TrimSpace(out.String()) != "[]" {
			t.Fatalf(
				"output = %q, want []",
				out.String(),
			)
		}
	})
}

type failAfterJSONWriter struct {
	buf   bytes.Buffer
	armed bool
	err   error
}

func (w *failAfterJSONWriter) Write(p []byte) (int, error) {
	if w.armed {
		return 0, w.err
	}

	n, err := w.buf.Write(p)
	if err != nil {
		return n, err
	}

	if bytes.Equal(
		bytes.TrimSpace(w.buf.Bytes()),
		[]byte("[]"),
	) {
		w.armed = true
	}

	return n, nil
}

func TestFinalizerJSON_NewlineWriteError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")

	out := &failAfterJSONWriter{
		err: wantErr,
	}

	var errOut bytes.Buffer

	f := finalizer{
		out:     out,
		errOut:  &errOut,
		colors:  newColorizer(true),
		jsonOut: true,
	}

	err := f.finalizeJSON([]fileOutcome{})

	if !errors.Is(err, wantErr) {
		t.Fatalf(
			"finalizeJSON() error = %v, want %v",
			err,
			wantErr,
		)
	}

	if !strings.Contains(
		errOut.String(),
		"failed to write json: write failed",
	) {
		t.Fatalf(
			"stderr = %q, want newline write error",
			errOut.String(),
		)
	}
}
