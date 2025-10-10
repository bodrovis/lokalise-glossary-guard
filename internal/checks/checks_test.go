package checks

import (
	"reflect"
	"testing"
)

// mockCheck implements Check for testing
type mockCheck struct {
	n  string
	p  int
	ff bool
	r  Result
}

func (m mockCheck) Name() string        { return m.n }
func (m mockCheck) Run(_ string) Result { return m.r }
func (m mockCheck) FailFast() bool      { return m.ff }
func (m mockCheck) Priority() int       { return m.p }

func TestSorted_EmptyRegistry(t *testing.T) {
	Reset()
	got := Sorted()
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d", len(got))
	}
}

func TestRegisterAndSorted_ByPriorityThenName(t *testing.T) {
	Reset()

	// Register in a messy order
	Register(mockCheck{n: "b", p: 1})
	Register(mockCheck{n: "c", p: 2})
	Register(mockCheck{n: "a", p: 1})

	got := Sorted()

	names := []string{got[0].Name(), got[1].Name(), got[2].Name()}
	want := []string{"a", "b", "c"} // p:1 (a,b) by name, then p:2 (c)
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("order mismatch\n got: %v\nwant: %v", names, want)
	}
}

func TestSorted_IsCopyNotAlias(t *testing.T) {
	Reset()

	Register(mockCheck{n: "x", p: 1})
	Register(mockCheck{n: "y", p: 2})

	sorted := Sorted()
	if len(sorted) != len(All) {
		t.Fatalf("length mismatch: sorted=%d all=%d", len(sorted), len(All))
	}

	// mutate the sorted slice element; All should be unaffected
	sorted[0] = mockCheck{n: "MUTATED", p: 0}

	if All[0].Name() == "MUTATED" {
		t.Fatalf("Sorted must return a copy; mutation leaked into All")
	}
}

func TestSorted_StableWhenNameAndPriorityEqual(t *testing.T) {
	Reset()

	// Two checks with the same priority and same name to verify stable sort
	first := mockCheck{n: "dup", p: 10}
	second := mockCheck{n: "dup", p: 10}

	Register(first)
	Register(second)

	got := Sorted()
	if len(got) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(got))
	}

	// Because the comparator treats them equal, SliceStable must preserve registration order.
	if got[0].Name() != "dup" || got[1].Name() != "dup" {
		t.Fatalf("unexpected names: %s, %s", got[0].Name(), got[1].Name())
	}

	// reflect the exact order by comparing pointer identity after type assertion
	g0, ok0 := got[0].(mockCheck)
	g1, ok1 := got[1].(mockCheck)
	if !ok0 || !ok1 {
		t.Fatalf("type assertion to mockCheck failed")
	}
	if !reflect.DeepEqual(g0, first) || !reflect.DeepEqual(g1, second) {
		t.Fatalf("stable order violated: got (%v, %v), want (%v, %v)", g0, g1, first, second)
	}
}

func TestSorted_SecondaryKeyName(t *testing.T) {
	Reset()

	// Same priority, different names → sorted lexicographically by Name()
	Register(mockCheck{n: "zeta", p: 5})
	Register(mockCheck{n: "alpha", p: 5})
	Register(mockCheck{n: "gamma", p: 5})

	got := Sorted()
	names := []string{got[0].Name(), got[1].Name(), got[2].Name()}
	want := []string{"alpha", "gamma", "zeta"}

	if !reflect.DeepEqual(names, want) {
		t.Fatalf("secondary sort by name failed\n got: %v\nwant: %v", names, want)
	}
}
