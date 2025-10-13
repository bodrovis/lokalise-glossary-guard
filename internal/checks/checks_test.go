package checks

import (
	"reflect"
	"sort"
	"testing"
)

// mockCheck implements Check for testing
type mockCheck struct {
	n  string
	p  int
	ff bool
	r  Result
}

func (m mockCheck) Name() string { return m.n }
func (m mockCheck) Run(_ []byte, _ string, _ []string) Result {
	if (m.r == Result{}) {
		return Passf(m.n, "ok")
	}
	return m.r
}
func (m mockCheck) FailFast() bool { return m.ff }
func (m mockCheck) Priority() int  { return m.p }

func names(xs []Check) []string {
	out := make([]string, len(xs))
	for i, c := range xs {
		out[i] = c.Name()
	}
	return out
}

func TestSplit_EmptyRegistry(t *testing.T) {
	Reset()
	crit, norm := Split()
	if len(crit) != 0 || len(norm) != 0 {
		t.Fatalf("expected both slices empty, got crit=%d norm=%d", len(crit), len(norm))
	}
}

func TestRegisterAndReplaceByName(t *testing.T) {
	Reset()

	// Register initial check
	Register(mockCheck{n: "dup", p: 10, ff: false})
	if len(All) != 1 {
		t.Fatalf("expected All to have 1 entry, got %d", len(All))
	}

	// Register with the same name should replace the previous one (not append)
	Register(mockCheck{n: "dup", p: 99, ff: true})
	if len(All) != 1 {
		t.Fatalf("expected replacement (len=1), got len=%d", len(All))
	}

	got := All[0].(mockCheck)
	if got.p != 99 || got.ff != true {
		t.Fatalf("replacement failed: got=%+v", got)
	}
}

func TestSplit_SortsCriticalByPriorityThenName(t *testing.T) {
	Reset()

	// Messy registration order
	Register(mockCheck{n: "z", p: 2, ff: true})
	Register(mockCheck{n: "a", p: 1, ff: true})
	Register(mockCheck{n: "b", p: 1, ff: true})
	Register(mockCheck{n: "m", p: 5, ff: true})

	crit, norm := Split()
	if len(norm) != 0 {
		t.Fatalf("expected no normal checks, got %d", len(norm))
	}
	if len(crit) != 4 {
		t.Fatalf("expected 4 critical checks, got %d", len(crit))
	}

	got := names(crit)
	want := []string{"a", "b", "z", "m"} // p:1 (a,b) by name, then p:2 (z), then p:5 (m)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("critical order mismatch\n got: %v\nwant: %v", got, want)
	}
}

func TestSplit_SplitsCriticalAndNormal(t *testing.T) {
	Reset()

	// Two critical, two normal
	Register(mockCheck{n: "crit-2", p: 2, ff: true})
	Register(mockCheck{n: "norm-x", p: 7, ff: false})
	Register(mockCheck{n: "crit-1", p: 1, ff: true})
	Register(mockCheck{n: "norm-y", p: 3, ff: false})

	crit, norm := Split()

	// Critical sorted by (priority, name)
	gotCrit := names(crit)
	wantCrit := []string{"crit-1", "crit-2"}
	if !reflect.DeepEqual(gotCrit, wantCrit) {
		t.Fatalf("critical order mismatch\n got: %v\nwant: %v", gotCrit, wantCrit)
	}

	// Normal: not guaranteed to be sorted by Split(); verify set equality
	gotNorm := names(norm)
	sort.Strings(gotNorm)
	wantNorm := []string{"norm-x", "norm-y"}
	if !reflect.DeepEqual(gotNorm, wantNorm) {
		t.Fatalf("normal membership mismatch\n got: %v\nwant: %v", gotNorm, wantNorm)
	}
}

func TestSplit_ReturnsCopies_NotAliases(t *testing.T) {
	Reset()
	Register(mockCheck{n: "a", p: 1})
	Register(mockCheck{n: "b", p: 2})

	crit, norm := Split()

	// Mutate the returned slices; global registry must not be affected
	if len(crit) > 0 {
		crit[0] = mockCheck{n: "MUTATED", p: 999}
	}
	if len(norm) > 0 {
		norm[0] = mockCheck{n: "MUTATED2", p: 999}
	}

	// Re-split and ensure names are unchanged
	crit2, norm2 := Split()
	got := append(names(crit2), names(norm2)...)
	want := []string{"a", "b"} // order across crit/norm is not important here
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registry was mutated via Split copies: got=%v want=%v", got, want)
	}
}

func TestReset_ClearsRegistry(t *testing.T) {
	Reset()
	Register(mockCheck{n: "x", p: 1})
	if len(All) != 1 {
		t.Fatalf("unexpected length before reset: %d", len(All))
	}
	Reset()
	if len(All) != 0 {
		t.Fatalf("reset did not clear registry; len=%d", len(All))
	}
}

func TestResultHelpers_StatusAndFormatting(t *testing.T) {
	r1 := Passf("pass", "ok")
	if r1.Name != "pass" || r1.Status != Pass || r1.Message != "ok" {
		t.Fatalf("Passf wrong: %+v", r1)
	}

	r2 := Warnf("warn", "heads-up %d", 42)
	if r2.Name != "warn" || r2.Status != Warn || r2.Message != "heads-up 42" {
		t.Fatalf("Warnf wrong: %+v", r2)
	}

	// без аргументов формат должен проходить как есть
	r3 := Failf("fail", "bad stuff")
	if r3.Name != "fail" || r3.Status != Fail || r3.Message != "bad stuff" {
		t.Fatalf("Failf wrong: %+v", r3)
	}

	r4 := Errorf("err", "boom: %s", "X")
	if r4.Name != "err" || r4.Status != Error || r4.Message != "boom: X" {
		t.Fatalf("Errorf wrong: %+v", r4)
	}
}

func TestFromFunc_ReturnsWarnAndMetadata(t *testing.T) {
	c := FromFunc("warn-check", false, 5, func(_ []byte, _ string, _ []string) Result {
		return Warnf("warn-check", "be careful")
	})

	if c.Name() != "warn-check" {
		t.Fatalf("Name mismatch: %s", c.Name())
	}
	if c.FailFast() {
		t.Fatalf("FailFast should be false")
	}
	if c.Priority() != 5 {
		t.Fatalf("Priority mismatch: %d", c.Priority())
	}

	out := c.Run(nil, "", nil)
	if out.Status != Warn || out.Name != "warn-check" || out.Message != "be careful" {
		t.Fatalf("Run result mismatch: %+v", out)
	}
}
