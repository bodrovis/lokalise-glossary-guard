// ensure_utf8_test.go
package checks

import "testing"

func TestEnsureUTF8_Metadata(t *testing.T) {
	c := ensureUTF8{}
	if got, want := c.Name(), "ensure-utf8-encoding"; got != want {
		t.Fatalf("Name() = %q, want %q", got, want)
	}
	if !c.FailFast() {
		t.Fatalf("FailFast() = false, want true")
	}
	if got, want := c.Priority(), 2; got != want {
		t.Fatalf("Priority() = %d, want %d", got, want)
	}
}

func TestEnsureUTF8_Run_Fail_EmptyFile(t *testing.T) {
	c := ensureUTF8{}
	res := c.Run([]byte(""), "", nil)
	if res.Status != Fail {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Fail, res.Message)
	}
	if res.Message == "" {
		t.Fatalf("expected non-empty message for empty input")
	}
}

func TestEnsureUTF8_Run_Pass_SimpleASCII(t *testing.T) {
	c := ensureUTF8{}
	res := c.Run([]byte("hello, world\n"), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Pass, res.Message)
	}
}

func TestEnsureUTF8_Run_Pass_UTF8BOM(t *testing.T) {
	c := ensureUTF8{}
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("with bom")...)
	res := c.Run(data, "", nil)
	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Pass, res.Message)
	}
}

func TestEnsureUTF8_Run_Pass_Multibyte(t *testing.T) {
	c := ensureUTF8{}
	res := c.Run([]byte("Привет, 你好!"), "", nil)
	if res.Status != Pass {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Pass, res.Message)
	}
}

func TestEnsureUTF8_Run_Fail_InvalidBytes(t *testing.T) {
	c := ensureUTF8{}

	// Build a buffer with invalid UTF-8 sequences.
	var bad []byte
	bad = append(bad, []byte("ok ")...) // valid prefix
	bad = append(bad, 0xFF, 0xFE)       // invalid
	bad = append(bad, ' ')              // separator
	bad = append(bad, 0xC3, 0x28)       // invalid continuation

	res := c.Run(bad, "", nil)
	if res.Status != Fail {
		t.Fatalf("Status = %s, want %s; msg=%q", res.Status, Fail, res.Message)
	}
	if res.Message == "" {
		t.Fatalf("expected error message with offset details")
	}
	// sanity: message should mention "invalid byte sequence" or "corruption"
	msg := res.Message
	if !containsLowerCheap(msg, "invalid byte sequence") && !containsLowerCheap(msg, "corruption") {
		t.Fatalf("unexpected message: %q", msg)
	}
}

/*** helpers ***/
func containsLowerCheap(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	S, Sub := []rune(s), []rune(sub)
	// cheap lower-case contains without extra imports
	lower := func(r rune) rune {
		if 'A' <= r && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}
	for i := range S {
		S[i] = lower(S[i])
	}
	for i := range Sub {
		Sub[i] = lower(Sub[i])
	}
	ls, lsub := string(S), string(Sub)
	return len(ls) >= len(lsub) && (len(lsub) == 0 || (func() bool {
		for i := 0; i+len(lsub) <= len(ls); i++ {
			if ls[i:i+len(lsub)] == lsub {
				return true
			}
		}
		return false
	})())
}
