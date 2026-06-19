package validate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard-core/pkg/checks"
	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestRenderResult(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

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
			name: "passed default",
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
			var b strings.Builder

			renderResult(&b, tt.resp)

			if b.String() != tt.want {
				t.Fatalf("output = %q, want %q", b.String(), tt.want)
			}
		})
	}
}

func TestRenderSummary(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	var b strings.Builder

	renderSummary(&b, "glossary.csv", guard.Summary{
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

func TestRenderEarlyExit_NoEarlyExit(t *testing.T) {
	var b strings.Builder

	renderEarlyExit(&b, guard.Summary{
		EarlyExit: false,
	})

	if b.String() != "" {
		t.Fatalf("output = %q, want empty", b.String())
	}
}

func TestRenderEarlyExit_PrintsMessage(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	total := len(checks.List())
	if total == 0 {
		t.Skip("no checks registered")
	}

	outcomes := make([]guard.Outcome, total-1)

	var b strings.Builder

	renderEarlyExit(&b, guard.Summary{
		EarlyExit:   true,
		EarlyCheck:  "headers",
		EarlyStatus: "FAIL",
		Outcomes:    outcomes,
	})

	want := `Stopped early due to fail-fast in check "headers" (FAIL). Skipped 1 remaining check(s).` + "\n"
	if b.String() != want {
		t.Fatalf("output = %q, want %q", b.String(), want)
	}
}

func TestRenderEarlyExit_DoesNotPrintNegativeSkippedCount(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	total := len(checks.List())
	outcomes := make([]guard.Outcome, total+10)

	var b strings.Builder

	renderEarlyExit(&b, guard.Summary{
		EarlyExit:   true,
		EarlyCheck:  "headers",
		EarlyStatus: "FAIL",
		Outcomes:    outcomes,
	})

	if !strings.Contains(b.String(), "Skipped 0 remaining check(s).") {
		t.Fatalf("output = %q, want skipped 0", b.String())
	}
}

func TestRenderCheckOutcomes(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	var b strings.Builder

	renderCheckOutcomes(&b, guard.Summary{
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
				Name:    "bad-delimiter",
				Status:  "FAIL",
				Message: "",
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

func TestRenderValidationReport(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	var b strings.Builder

	renderValidationReport(&b, "glossary.csv", guard.Summary{
		Pass: 1,
		Warn: 1,
		Fail: 0,
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
			t.Fatalf("output = %q, want it to contain %q", got, want)
		}
	}
}

func TestRenderFileHeader_FirstFile(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	var b strings.Builder

	opts := checks.RunOptions{
		FixMode:       checks.FixNone,
		RerunAfterFix: true,
		HardFailOnErr: false,
	}

	renderFileHeader(&b, 0, "glossary.csv", "---", opts)

	want := "" +
		"---\n" +
		"Validating: glossary.csv\n" +
		"---\n\n" +
		fmt.Sprintf("Mode: FixMode=%v, RerunAfterFix=true, HardFailOnErr=false\n\n", checks.FixNone)

	if b.String() != want {
		t.Fatalf("output = %q, want %q", b.String(), want)
	}
}

func TestRenderFileHeader_NextFileStartsWithBlankLine(t *testing.T) {
	oldNoColor := noColor
	t.Cleanup(func() {
		noColor = oldNoColor
	})
	noColor = true

	var b strings.Builder

	renderFileHeader(&b, 1, "second.csv", "---", checks.RunOptions{})

	if !strings.HasPrefix(b.String(), "\n---\n") {
		t.Fatalf("output = %q, want leading blank line before separator", b.String())
	}
}

func TestOneLine(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"hello", "hello"},
		{" hello   world ", "hello world"},
		{"hello\nworld", "hello world"},
		{"hello\r\nworld", "hello world"},
		{"hello\n\n\tworld   again", "hello world again"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.in), func(t *testing.T) {
			got := oneLine(tt.in)
			if got != tt.want {
				t.Fatalf("oneLine(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
