// ensure_unique_headers_test.go
package checks

import "testing"

func TestEnsureUniqueHeaders_Metadata(t *testing.T) {
	c := ensureUniqueHeaders{}
	if c.Name() != "ensure-unique-headers" {
		t.Fatalf("Name() = %q", c.Name())
	}
	if c.FailFast() {
		t.Fatalf("FailFast() should be false")
	}
	if c.Priority() != 102 {
		t.Fatalf("Priority() = %d, want 102", c.Priority())
	}
}

func TestEnsureUniqueHeaders_Error_EmptyFile(t *testing.T) {
	c := ensureUniqueHeaders{}
	res := c.Run([]byte(""), "", nil)
	if res.Status != Error || !containsLower(res.Message, "cannot read header: empty file") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueHeaders_Pass_AllUnique(t *testing.T) {
	c := ensureUniqueHeaders{}
	content := "term;description;casesensitive;translatable;forbidden;tags;en;de_DE\n" +
		"k;d;no;yes;no;t1,t2;hello;hallo\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueHeaders_Fail_SimpleDuplicate(t *testing.T) {
	c := ensureUniqueHeaders{}
	content := "term;description;en;en\n" +
		"a;b;c;d\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail || !containsLower(res.Message, "duplicate header name(s) found") {
		t.Fatalf("Status=%s msg=%q", res.Status, res.Message)
	}
	// should mention columns indexes
	if !containsLower(res.Message, "columns 3 and 4") {
		t.Fatalf("Expected column positions in message: %q", res.Message)
	}
}

func TestEnsureUniqueHeaders_Fail_CaseInsensitiveDuplicate(t *testing.T) {
	c := ensureUniqueHeaders{}
	content := "Term;term;description\n" +
		"x;y;z\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueHeaders_Fail_HyphenUnderscoreConsideredSame(t *testing.T) {
	c := ensureUniqueHeaders{}
	// en-US vs en_US should be treated as duplicate
	content := "en-US;en_US;term;description\n" +
		"a;b;c;d\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueHeaders_Fail_SpacesTrimmed(t *testing.T) {
	c := ensureUniqueHeaders{}
	// " term " vs "term" => duplicate after trim
	content := "  term  ;term;description\n" +
		"a;b;c\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueHeaders_Pass_WithBOM(t *testing.T) {
	c := ensureUniqueHeaders{}
	content := "\uFEFFterm;description;en\n" +
		"t;d;hello\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status=%s want PASS (BOM stripped), msg=%q", res.Status, res.Message)
	}
}

func TestEnsureUniqueHeaders_Fail_MultipleDuplicates_TruncatesList(t *testing.T) {
	c := ensureUniqueHeaders{}
	// Build a header with >10 duplicate pairs to trigger truncation
	// We'll repeat "x" many times among unique others.
	header := "term;description"
	for i := 0; i < 12; i++ { // 12 occurrences of x -> 11 duplicate reports (but message truncates to 10 shown)
		if i == 0 {
			header += ";x"
		} else {
			header += ";x"
		}
	}
	header += "\n" + "a;b"
	for i := 0; i < 12; i++ {
		header += ";v"
	}
	header += "\n"

	res := c.Run([]byte(header), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
	if !containsLower(res.Message, "...and ") {
		t.Fatalf("Expected truncation hint ('...and N more') in message, got: %q", res.Message)
	}
	if !containsLower(res.Message, "duplicate header name(s) found") {
		t.Fatalf("Expected duplicate header headline, got: %q", res.Message)
	}
}

func TestEnsureUniqueHeaders_Fail_DifferentSpellingSameNormalized(t *testing.T) {
	c := ensureUniqueHeaders{}
	// "EN_Description" and "en-description" normalize to same key
	content := "term;description;EN_Description;en-description\n" +
		"k;d;foo;bar\n"
	res := c.Run([]byte(content), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status=%s want FAIL, msg=%q", res.Status, res.Message)
	}
}
