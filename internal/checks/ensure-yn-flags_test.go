// ensure_yn_flags_test.go
package checks

import (
	"strings"
	"testing"
)

func TestEnsureYNFlags_Metadata(t *testing.T) {
	c := ensureYNFlags{}
	if c.Name() != "ensure-yn-flags" {
		t.Fatalf("Name() = %q", c.Name())
	}
	if c.FailFast() {
		t.Fatalf("FailFast() should be false")
	}
	if c.Priority() != 100 {
		t.Fatalf("Priority() = %d, want 100", c.Priority())
	}
}

func TestEnsureYNFlags_Error_EmptyFile(t *testing.T) {
	c := ensureYNFlags{}
	res := c.Run([]byte(""), "", nil)
	if res.Status != Error || !containsLower(res.Message, "cannot read header: empty file") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureYNFlags_Pass_NoFlagColumns(t *testing.T) {
	c := ensureYNFlags{}
	content := "term;description;tags\n" +
		"t;d;a,b\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass || !containsLower(res.Message, "no y/n flag columns") {
		t.Fatalf("Status=%s want PASS (no flags), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureYNFlags_Pass_AllValidYesNo_MixedCase(t *testing.T) {
	c := ensureYNFlags{}
	content := "term;description;casesensitive;translatable;forbidden\n" +
		"k;d;YES;no;No\n" +
		"x;y;yes;NO;yes\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (case-insensitive yes/no), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureYNFlags_Fail_EmptyValues(t *testing.T) {
	c := ensureYNFlags{}
	content := "term;description;casesensitive;translatable;forbidden\n" +
		"a;b;;no;yes\n" + // empty casesensitive
		"c;d;yes; ;no\n" // whitespace translatable
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "invalid values in y/n flag columns") {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
	// Should mention which columns and lines
	if !containsLower(res.Message, "line 2, column casesensitive") ||
		!containsLower(res.Message, "line 3, column translatable") {
		t.Fatalf("Expected line/column details in message: %q", res.Message)
	}
}

func TestEnsureYNFlags_Fail_InvalidTokens(t *testing.T) {
	c := ensureYNFlags{}
	content := "term;description;casesensitive;translatable\n" +
		"a;b;true;no\n" +
		"c;d;yes;0\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL (true/0 not allowed), msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "true") || !containsLower(res.Message, "0") {
		t.Fatalf("Expected to include offending values: %q", res.Message)
	}
}

func TestEnsureYNFlags_Error_CSVParseError_UnbalancedQuote(t *testing.T) {
	c := ensureYNFlags{}
	content := "term;description;casesensitive\n" +
		`x;y;"no` + "\n" // unbalanced quotes -> CSV parse error at line 2
	res := c.Run([]byte(content), "", nil)
	if res.Status != Error || !containsLower(res.Message, "csv parse error at line 2") {
		t.Fatalf("Status=%s want ERROR with line note, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureYNFlags_Fail_SortedByColumnThenLine(t *testing.T) {
	c := ensureYNFlags{}
	// Make violations in two columns out of order: we expect lexicographic by column, then by line.
	content := "term;description;forbidden;casesensitive\n" +
		"a;b;bad;yes\n" + // line 2, forbidden invalid
		"c;d;no;bad\n" + // line 3, casesensitive invalid
		"e;f;BAD;bad\n" // line 4, forbidden invalid AND casesensitive invalid
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
	// The first listed should be "casesensitive" line 3 (column name 'casesensitive' < 'forbidden')
	// Then "casesensitive" line 4, then "forbidden" line 2, then "forbidden" line 4
	msg := strings.ToLower(res.Message)
	firstIdx := strings.Index(msg, "line 3, column casesensitive")
	if firstIdx == -1 {
		t.Fatalf("Expected first to mention 'line 3, column casesensitive'. Got: %q", res.Message)
	}
	// Ensure order constraint roughly holds by checking the relative order of key fragments
	idxC4 := strings.Index(msg, "line 4, column casesensitive")
	idxF2 := strings.Index(msg, "line 2, column forbidden")
	idxF4 := strings.Index(msg, "line 4, column forbidden")
	if firstIdx >= idxC4 || idxC4 >= idxF2 || idxF2 >= idxF4 {
		t.Fatalf("Unexpected ordering in message: %q", res.Message)
	}
}

func TestEnsureYNFlags_Fail_TruncatesLongList(t *testing.T) {
	c := ensureYNFlags{}
	// Build >10 invalids to trigger truncation
	content := "term;description;forbidden\n"
	for i := 0; i < 12; i++ {
		content += "t;d;maybe\n"
	}
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "...and ") {
		t.Fatalf("Expected truncation hint ('...and N more'), got: %q", res.Message)
	}
}
