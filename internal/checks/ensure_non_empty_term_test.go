// ensure_non_empty_term_test.go
package checks

import "testing"

func TestEnsureNonEmptyTerm_Metadata(t *testing.T) {
	c := ensureNonEmptyTerm{}
	if c.Name() != "ensure-non-empty-term" {
		t.Fatalf("Name() = %q", c.Name())
	}
	if c.FailFast() {
		t.Fatalf("FailFast() should be false")
	}
	if c.Priority() != 100 {
		t.Fatalf("Priority() = %d, want 100", c.Priority())
	}
}

func TestEnsureNonEmptyTerm_Fail_EmptyTermValue(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\n;no term here\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "term value is required") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Pass_Minimal(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\nfoo;bar\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Pass_EmptyDescriptionAllowed(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\nfoo;\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Fail_WhitespaceOnly(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\n   ;desc\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "term value is required") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Fail_QuotedEmpty(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\n\"\";no term here\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "term value is required") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Pass_TermWithSpacesAround(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\n  foo  ;desc\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Pass_SemicolonsCommasInsideQuotes(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\n\"a;b,c\";\"x;y,z\"\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Fail_EmptyOnSecondRow_WithLineNumber(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\nok;d1\n;d2\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "term value is required") || !containsLower(res.Message, "line 3") {
		t.Fatalf("Status=%s msg=%q (expect line number 3)", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Pass_TermColumnNotFirst(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "description;term;tags\ndesc;foo;bar\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Error_NoTermColumn(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "description;tags\nx;y\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Error || !containsLower(res.Message, "header does not contain 'term'") {
		t.Fatalf("Status=%s want ERROR, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Pass_CRLF(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\r\nalpha;beta\r\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureNonEmptyTerm_Fail_MissingBeforeDelimiter(t *testing.T) {
	c := ensureNonEmptyTerm{}
	content := "term;description\n;desc\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "term value is required") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}
