package validate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func newTestRenderer(b *strings.Builder) fileRenderer {
	return fileRenderer{
		out:       b,
		separator: "---",
		colors:    newColorizer(true),
	}
}

func TestFileRenderer_Result(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp guard.ValidateResponse
		want string
	}{
		{
			name: "failed",
			resp: guard.ValidateResponse{Failed: true},
			want: "Result: FAILED\n",
		},
		{
			name: "warned",
			resp: guard.ValidateResponse{Warned: true},
			want: "Result: PASSED WITH WARNINGS\n",
		},
		{
			name: "passed",
			resp: guard.ValidateResponse{Passed: true},
			want: "Result: PASSED\n",
		},
		{
			name: "zero response defaults to passed",
			resp: guard.ValidateResponse{},
			want: "Result: PASSED\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var b strings.Builder
			r := newTestRenderer(&b)

			r.result(tt.resp)

			if b.String() != tt.want {
				t.Fatalf(
					"output = %q, want %q",
					b.String(),
					tt.want,
				)
			}
		})
	}
}

func TestFileRenderer_Summary(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	r := newTestRenderer(&b)

	r.summary("glossary.csv", guard.Summary{
		Pass:   10,
		Warn:   2,
		Fail:   1,
		Errors: 3,
	})

	want := "\nSummary for glossary.csv: 10 passed, 2 warning(s), 1 failed, 3 errors\n"

	if b.String() != want {
		t.Fatalf("output = %q, want %q", b.String(), want)
	}
}

func TestFileRenderer_EarlyExit(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		var b strings.Builder
		r := newTestRenderer(&b)

		r.earlyExit(guard.Summary{})

		if b.Len() != 0 {
			t.Fatalf("output = %q, want empty", b.String())
		}
	})

	t.Run("prints skipped checks", func(t *testing.T) {
		total := len(checks.List())
		if total == 0 {
			t.Skip("no checks registered")
		}

		var b strings.Builder
		r := newTestRenderer(&b)

		r.earlyExit(guard.Summary{
			EarlyExit:   true,
			EarlyCheck:  "headers",
			EarlyStatus: "FAIL",
			Outcomes:    make([]guard.Outcome, total-1),
		})

		want := `Stopped early due to fail-fast in check "headers" (FAIL). Skipped 1 remaining check(s).` + "\n"

		if b.String() != want {
			t.Fatalf("output = %q, want %q", b.String(), want)
		}
	})

	t.Run("never prints negative skipped count", func(t *testing.T) {
		var b strings.Builder
		r := newTestRenderer(&b)

		r.earlyExit(guard.Summary{
			EarlyExit:   true,
			EarlyCheck:  "headers",
			EarlyStatus: "FAIL",
			Outcomes:    make([]guard.Outcome, len(checks.List())+10),
		})

		if !strings.Contains(
			b.String(),
			"Skipped 0 remaining check(s).",
		) {
			t.Fatalf(
				"output = %q, want skipped 0",
				b.String(),
			)
		}
	})
}

func TestFileRenderer_CheckOutcomes(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	r := newTestRenderer(&b)

	r.checkOutcomes(guard.Summary{
		Outcomes: []guard.Outcome{
			{
				Name:     "headers",
				Status:   "PASS",
				Message:  "all good",
				Critical: true,
			},
			{
				Name:    "empty-lines",
				Status:  "WARN",
				Message: "line 1\r\nline 2\n\nline 3",
				Changed: true,
				Note:    "removed\nempty line",
			},
			{
				Name:   "bad-delimiter",
				Status: "FAIL",
			},
		},
	})

	want := "" +
		"→ [CRIT] headers ... PASS\n" +
		"   all good\n" +
		"→ [NORM] empty-lines ... WARN [changed]\n" +
		"   line 1 line 2 line 3 | note: removed empty line\n" +
		"→ [NORM] bad-delimiter ... FAIL\n" +
		"   -\n"

	if b.String() != want {
		t.Fatalf("output = %q, want %q", b.String(), want)
	}
}

func TestFileRenderer_ValidationReport(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	r := newTestRenderer(&b)

	r.validationReport("glossary.csv", guard.Summary{
		Pass: 1,
		Warn: 1,
		Outcomes: []guard.Outcome{
			{
				Name:    "headers",
				Status:  "PASS",
				Message: "ok",
			},
			{
				Name:    "empty-lines",
				Status:  "WARN",
				Message: "has empty line",
			},
		},
	})

	got := b.String()

	for _, want := range []string{
		"→ [NORM] headers ... PASS\n",
		"   ok\n",
		"→ [NORM] empty-lines ... WARN\n",
		"   has empty line\n",
		"Summary for glossary.csv: 1 passed, 1 warning(s), 0 failed, 0 errors\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf(
				"output = %q, want it to contain %q",
				got,
				want,
			)
		}
	}
}

func TestFileRenderer_FileHeader(t *testing.T) {
	t.Parallel()

	t.Run("first file", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder

		opts := checks.RunOptions{
			FixMode:       checks.FixNone,
			RerunAfterFix: true,
			HardFailOnErr: false,
		}

		r := fileRenderer{
			out:       &b,
			separator: "---",
			options:   opts,
			colors:    newColorizer(true),
		}

		r.fileHeader(0, "glossary.csv")

		want := "" +
			"---\n" +
			"Validating: glossary.csv\n" +
			"---\n\n" +
			fmt.Sprintf(
				"Mode: FixMode=%v, RerunAfterFix=true, HardFailOnErr=false\n\n",
				checks.FixNone,
			)

		if b.String() != want {
			t.Fatalf("output = %q, want %q", b.String(), want)
		}
	})

	t.Run("subsequent file starts with blank line", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		r := newTestRenderer(&b)

		r.fileHeader(1, "second.csv")

		if !strings.HasPrefix(b.String(), "\n---\n") {
			t.Fatalf(
				"output = %q, want leading blank line",
				b.String(),
			)
		}
	})
}

func TestOneLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"unchanged", "hello", "hello"},
		{"spaces", " hello   world ", "hello world"},
		{"newline", "hello\nworld", "hello world"},
		{"crlf", "hello\r\nworld", "hello world"},
		{
			"mixed whitespace",
			"hello\n\n\tworld   again",
			"hello world again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := oneLine(tt.in)

			if got != tt.want {
				t.Fatalf(
					"oneLine(%q) = %q, want %q",
					tt.in,
					got,
					tt.want,
				)
			}
		})
	}
}
