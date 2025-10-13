package checks

import "testing"

func TestEnsureUniqueTerms_Metadata(t *testing.T) {
	c := ensureUniqueTermsCS{}
	if c.Name() != "ensure-unique-terms" {
		t.Fatalf("Name() = %q", c.Name())
	}
	if c.FailFast() {
		t.Fatalf("FailFast() should be false")
	}
	if c.Priority() != 100 {
		t.Fatalf("Priority() = %d, want 100", c.Priority())
	}
}

func TestEnsureUniqueTerms_Error_EmptyFile(t *testing.T) {
	c := ensureUniqueTermsCS{}
	res := c.Run([]byte(""), "", nil)
	if res.Status != Error || !containsLower(res.Message, "cannot read header: empty file") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Error_NoTermColumn(t *testing.T) {
	c := ensureUniqueTermsCS{}
	content := "description;tags\nx;y\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Error || !containsLower(res.Message, "header does not contain 'term'") {
		t.Fatalf("Status=%s want ERROR, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Pass_AllUnique_Minimal(t *testing.T) {
	c := ensureUniqueTermsCS{}
	content := "term;description\nfoo;bar\nbaz;qux\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Pass_CaseSensitiveDifferentValues(t *testing.T) {
	c := ensureUniqueTermsCS{}
	// "User" and "user" are NOT duplicates (case-sensitive)
	content := "term;description\nUser;A\nuser;B\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (case-sensitive), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Warn_SimpleDuplicate(t *testing.T) {
	c := ensureUniqueTermsCS{}
	content := "term;description\ndup;one\ndup;two\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Warn || !containsLower(res.Message, "duplicate terms") {
		t.Fatalf("Status=%s want WARN, msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "lines 2 and 3") {
		t.Fatalf("Expected line numbers 2 and 3 in message: %q", res.Message)
	}
}

func TestEnsureUniqueTerms_Warn_TrimmedDuplicates(t *testing.T) {
	c := ensureUniqueTermsCS{}
	// " foo " and "foo" should be considered duplicates after TrimSpace
	content := "term;description\n  foo  ;d1\nfoo;d2\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Warn {
		t.Fatalf("Status=%s want WARN (trimmed duplicates), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Pass_EmptyTermSkipped(t *testing.T) {
	c := ensureUniqueTermsCS{}
	// empty term row is ignored by this check (handled by ensure-non-empty-term)
	content := "term;description\nfoo;d1\n;d2\nbar;d3\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (empty term ignored here), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Pass_TermColumnNotFirst(t *testing.T) {
	c := ensureUniqueTermsCS{}
	content := "description;term;tags\ndesc;foo;bar\nx;bar;y\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (term not first), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Pass_CRLF(t *testing.T) {
	c := ensureUniqueTermsCS{}
	content := "term;description\r\nalpha;beta\r\nbeta;gamma\r\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (CRLF), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Warn_TruncatesLongDuplicateList(t *testing.T) {
	c := ensureUniqueTermsCS{}
	// 1 header + 1 first occurrence + 12 duplicates => 12 hits, message should truncate to 10 shown
	content := "term;description\nx;d0\n"
	for i := 0; i < 12; i++ {
		content += "x;d\n"
	}
	res := c.Run([]byte(content), "", nil)
	if res.Status != Warn {
		t.Fatalf("Status=%s want WARN, msg=%q", res.Status, res.Message)
	}

	if !containsLower(res.Message, "...and ") {
		t.Fatalf("Expected truncation hint in message, got: %q", res.Message)
	}
}

func TestEnsureUniqueTerms_Warn_BOMInFirstDataCellThenSameValue(t *testing.T) {
	c := ensureUniqueTermsCS{}
	// First data row has BOM+foo (trimmed on line==2), second row "foo" -> duplicate
	content := "term;description\n\uFEFFfoo;d1\nfoo;d2\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Warn {
		t.Fatalf("Status=%s want WARN (BOM normalized), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueTerms_Error_CSVParseError(t *testing.T) {
	c := ensureUniqueTermsCS{}

	content := "term;description\n\"oops;bad\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Error || !containsLower(res.Message, "csv parse error at line") {
		t.Fatalf("Status=%s want ERROR with parse note, msg=%q", res.Status, res.Message)
	}
}
