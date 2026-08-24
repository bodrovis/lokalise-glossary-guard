//go:build js && wasm

package main

import (
	"encoding/json/v2"
	"strings"
	"syscall/js"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestValidateGlossaryGuardEnvelope_MissingInput(t *testing.T) {
	env := validateGlossaryGuardEnvelope(nil)

	if env.OK {
		t.Fatal("OK = true, want false")
	}

	if env.Result != nil {
		t.Fatalf("Result = %#v, want nil", env.Result)
	}

	if env.Error != "missing input" {
		t.Fatalf("Error = %q, want %q", env.Error, "missing input")
	}
}

func TestValidateGlossaryGuardEnvelope_InvalidJSONString(t *testing.T) {
	env := validateGlossaryGuardEnvelope([]js.Value{
		js.ValueOf("not json"),
	})

	if env.OK {
		t.Fatal("OK = true, want false")
	}

	if env.Result != nil {
		t.Fatalf("Result = %#v, want nil", env.Result)
	}

	if !strings.Contains(env.Error, "invalid input json") {
		t.Fatalf("Error = %q, want invalid input json error", env.Error)
	}
}

func TestValidateGlossaryGuardEnvelope_ValidObjectInput(t *testing.T) {
	env := validateGlossaryGuardEnvelope([]js.Value{
		js.ValueOf(map[string]any{
			"path": "glossary.csv",
			"data": validWASMCSV(),
		}),
	})

	if !env.OK {
		t.Fatalf("OK = false, want true; error: %q", env.Error)
	}

	if env.Result == nil {
		t.Fatal("Result = nil, want validation result")
	}

	if env.Result.Status != guard.StatusPassed {
		t.Fatalf("Status = %q, want %q; result: %#v", env.Result.Status, guard.StatusPassed, env.Result)
	}

	if !env.Result.Passed {
		t.Fatalf("Passed = false, want true; result: %#v", env.Result)
	}

	if env.Result.Path != "glossary.csv" {
		t.Fatalf("Path = %q, want %q", env.Result.Path, "glossary.csv")
	}
}

func TestValidateGlossaryGuardEnvelope_ValidJSONStringInput(t *testing.T) {
	input := `{"path":"glossary.csv","data":` + mustJSONQuote(t, validWASMCSV()) + `}`

	env := validateGlossaryGuardEnvelope([]js.Value{
		js.ValueOf(input),
	})

	if !env.OK {
		t.Fatalf("OK = false, want true; error: %q", env.Error)
	}

	if env.Result == nil {
		t.Fatal("Result = nil, want validation result")
	}

	if env.Result.Status != guard.StatusPassed {
		t.Fatalf("Status = %q, want %q; result: %#v", env.Result.Status, guard.StatusPassed, env.Result)
	}
}

func TestValidateGlossaryGuardEnvelope_ValidationFailureIsNotWASMFailure(t *testing.T) {
	env := validateGlossaryGuardEnvelope([]js.Value{
		js.ValueOf(map[string]any{
			"path": "glossary.csv",
			"data": "term,description,en\nsession,A login session,session\n",
		}),
	})

	if !env.OK {
		t.Fatalf("OK = false, want true; error: %q", env.Error)
	}

	if env.Error != "" {
		t.Fatalf("Error = %q, want empty", env.Error)
	}

	if env.Result == nil {
		t.Fatal("Result = nil, want validation result")
	}

	if env.Result.Status != guard.StatusFailed {
		t.Fatalf(
			"Status = %q, want %q; result: %#v",
			env.Result.Status,
			guard.StatusFailed,
			env.Result,
		)
	}

	if !env.Result.Failed {
		t.Fatalf("Failed = false, want true; result: %#v", env.Result)
	}
}

func TestValidateGlossaryGuard_ReturnsEncodedEnvelope(t *testing.T) {
	raw := validateGlossaryGuard(js.Undefined(), []js.Value{
		js.ValueOf(map[string]any{
			"path": "glossary.csv",
			"data": validWASMCSV(),
		}),
	})

	env := decodeWASMEnvelope(t, raw)

	if !env.OK {
		t.Fatalf("OK = false, want true; error: %q", env.Error)
	}

	if env.Result == nil {
		t.Fatal("Result = nil, want validation result")
	}

	if env.Result.Status != guard.StatusPassed {
		t.Fatalf("Status = %q, want %q; result: %#v", env.Result.Status, guard.StatusPassed, env.Result)
	}
}

func TestValidateGlossaryGuard_RecoversFromPanic(t *testing.T) {
	obj := js.Global().Get("Object").New()
	obj.Set("self", obj)

	raw := validateGlossaryGuard(js.Undefined(), []js.Value{obj})
	env := decodeWASMEnvelope(t, raw)

	if env.OK {
		t.Fatal("OK = true, want false")
	}

	if env.Result != nil {
		t.Fatalf("Result = %#v, want nil", env.Result)
	}

	if !strings.Contains(env.Error, "wasm panic") {
		t.Fatalf("Error = %q, want wasm panic error", env.Error)
	}
}

func TestValidateGlossaryGuard_ReturnsEncodedErrorEnvelope(t *testing.T) {
	raw := validateGlossaryGuard(
		js.Undefined(),
		[]js.Value{js.ValueOf("not json")},
	)

	env := decodeWASMEnvelope(t, raw)

	if env.OK {
		t.Fatal("OK = true, want false")
	}

	if env.Result != nil {
		t.Fatalf("Result = %#v, want nil", env.Result)
	}

	if !strings.Contains(env.Error, "invalid input json") {
		t.Fatalf(
			"Error = %q, want invalid input json error",
			env.Error,
		)
	}
}

func decodeWASMEnvelope(t *testing.T, raw any) wasmValidateEnvelope {
	t.Helper()

	s, ok := raw.(string)
	if !ok {
		t.Fatalf("raw response has type %T, want string", raw)
	}

	var env wasmValidateEnvelope
	if err := json.Unmarshal([]byte(s), &env); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nraw: %s", err, s)
	}

	return env
}

func mustJSONQuote(t *testing.T, s string) string {
	t.Helper()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	return string(data)
}

func validWASMCSV() string {
	return "term;description;en;fr\nsession;A login session;session;session\n"
}
