package validate

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestApplyGuardResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		resp        guard.ValidateResponse
		wantPassed  int
		wantWarned  int
		wantFailed  int
		wantErrored int
		wantValFail bool
	}{
		{
			name: "passed",
			resp: guard.ValidateResponse{
				Passed:  true,
				Summary: guard.Summary{FilePath: "ok.csv", Pass: 10},
			},
			wantPassed: 1,
		},
		{
			name: "warned",
			resp: guard.ValidateResponse{
				Warned:  true,
				Summary: guard.Summary{FilePath: "warn.csv", Warn: 2},
			},
			wantWarned: 1,
		},
		{
			name: "failed",
			resp: guard.ValidateResponse{
				Failed:  true,
				Summary: guard.Summary{FilePath: "bad.csv", Fail: 1},
			},
			wantFailed:  1,
			wantValFail: true,
		},
		{
			name: "errored",
			resp: guard.ValidateResponse{
				Failed:  true,
				Errored: true,
				Summary: guard.Summary{
					FilePath: "errored.csv",
					Errors:   1,
				},
			},
			wantFailed:  1,
			wantErrored: 1,
			wantValFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got fileOutcome
			applyGuardResponse(&got, tt.resp)

			if got.Passed != tt.wantPassed {
				t.Fatalf("Passed = %d, want %d", got.Passed, tt.wantPassed)
			}
			if got.Warned != tt.wantWarned {
				t.Fatalf("Warned = %d, want %d", got.Warned, tt.wantWarned)
			}
			if got.Failed != tt.wantFailed {
				t.Fatalf("Failed = %d, want %d", got.Failed, tt.wantFailed)
			}
			if got.Errored != tt.wantErrored {
				t.Fatalf("Errored = %d, want %d", got.Errored, tt.wantErrored)
			}
			if got.HadValFail != tt.wantValFail {
				t.Fatalf(
					"HadValFail = %v, want %v",
					got.HadValFail,
					tt.wantValFail,
				)
			}
			if got.Summary == nil {
				t.Fatal("Summary = nil, want summary")
			}
		})
	}
}

func TestFileRunConfig_OpErrorOutcome(t *testing.T) {
	t.Parallel()

	cfg := fileRunConfig{
		Separator: "---",
		Colors:    newColorizer(true),
	}

	var b strings.Builder

	oc := cfg.opErrorOutcome(
		7,
		"missing.csv",
		&b,
		errors.New("open failed"),
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

	const want = "ERROR: open failed\n---\n"

	if oc.Output != want {
		t.Fatalf("Output = %q, want %q", oc.Output, want)
	}
	if b.String() != want {
		t.Fatalf("builder = %q, want %q", b.String(), want)
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

func TestAggregateError(t *testing.T) {
	t.Parallel()

	const opErr = "one or more files could not be validated due to an error"

	tests := []struct {
		name string
		agg  aggregateResult
		want string
	}{
		{
			name: "no error",
			agg:  aggregateResult{},
		},
		{
			name: "validation failure",
			agg: aggregateResult{
				HadValFail: true,
			},
			want: "validation failed",
		},
		{
			name: "operational error",
			agg: aggregateResult{
				HadOpErr: true,
			},
			want: opErr,
		},
		{
			name: "operational error takes precedence",
			agg: aggregateResult{
				HadOpErr:   true,
				HadValFail: true,
			},
			want: opErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := aggregateError(tt.agg)

			if tt.want == "" {
				if err != nil {
					t.Fatalf("aggregateError() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("aggregateError() error = nil, want %q", tt.want)
			}
			if err.Error() != tt.want {
				t.Fatalf(
					"aggregateError() error = %q, want %q",
					err.Error(),
					tt.want,
				)
			}
		})
	}
}

func TestFinalizer_PrintAndAggregateSingleFile(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	f := finalizer{
		out:    &out,
		colors: newColorizer(true),
	}

	got := f.printAndAggregate(
		[]fileOutcome{
			{
				Path:   "ok.csv",
				Output: "file output\n",
				Passed: 1,
				Summary: &guard.Summary{
					Warn: 2,
				},
			},
		},
		1,
		time.Now(),
	)

	want := aggregateResult{
		Passed: 1,
		Warns:  2,
	}

	if got != want {
		t.Fatalf("aggregate = %#v, want %#v", got, want)
	}

	output := out.String()

	if !strings.Contains(output, "file output\n") {
		t.Fatalf("output = %q, want file output", output)
	}
	if strings.Contains(output, "Overall:") {
		t.Fatalf("output = %q, want no overall line", output)
	}
	if !strings.Contains(output, "Total time:") {
		t.Fatalf("output = %q, want total time", output)
	}
}

func TestFinalizer_PrintAndAggregateMultipleFiles(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	f := finalizer{
		out:    &out,
		colors: newColorizer(true),
	}

	got := f.printAndAggregate(
		[]fileOutcome{
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
		},
		3,
		time.Now(),
	)

	want := aggregateResult{
		HadOpErr:   true,
		HadValFail: true,
		Passed:     1,
		Warns:      3,
		Failed:     1,
		Errored:    1,
	}

	if got != want {
		t.Fatalf("aggregate = %#v, want %#v", got, want)
	}

	output := out.String()

	for _, want := range []string{
		"ok output\n",
		"bad output\n",
		"missing output\n",
		"Overall: 1 passed, 3 warning(s), 1 failed, 1 error(s)",
		"Total time:",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf(
				"output = %q, want it to contain %q",
				output,
				want,
			)
		}
	}
}
