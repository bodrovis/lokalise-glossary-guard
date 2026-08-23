//go:build js && wasm

package main

import (
	"encoding/json/v2"
	"testing"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func TestErrorEnvelope(t *testing.T) {
	env := errorEnvelope("boom")

	if env.OK {
		t.Fatal("OK = true, want false")
	}

	if env.Error != "boom" {
		t.Fatalf("Error = %q, want %q", env.Error, "boom")
	}

	if env.Result != nil {
		t.Fatalf("Result = %#v, want nil", env.Result)
	}
}

func TestEncodeEnvelope_Error(t *testing.T) {
	raw := encodeEnvelope(errorEnvelope("missing input"))

	var body map[string]any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nraw: %s", err, raw)
	}

	if body["ok"] != false {
		t.Fatalf("ok = %#v, want false", body["ok"])
	}

	if body["error"] != "missing input" {
		t.Fatalf("error = %#v, want %q", body["error"], "missing input")
	}

	if _, ok := body["result"]; ok {
		t.Fatalf("result is present, want omitted: %s", raw)
	}
}

func TestEncodeEnvelope_Success(t *testing.T) {
	raw := encodeEnvelope(wasmValidateEnvelope{
		OK: true,
		Result: &guard.ValidateResponse{
			Path:      "glossary.csv",
			Status:    guard.StatusPassed,
			Passed:    true,
			FixedData: []byte("must not be serialized"),
		},
	})

	var body map[string]any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nraw: %s", err, raw)
	}

	if body["ok"] != true {
		t.Fatalf("ok = %#v, want true", body["ok"])
	}

	if _, ok := body["error"]; ok {
		t.Fatalf("error is present, want omitted: %s", raw)
	}

	result, ok := body["result"].(map[string]any)
	if !ok {
		t.Fatalf("result = %#v, want object", body["result"])
	}

	if result["path"] != "glossary.csv" {
		t.Fatalf("result.path = %#v, want %q", result["path"], "glossary.csv")
	}

	if result["status"] != string(guard.StatusPassed) {
		t.Fatalf("result.status = %#v, want %q", result["status"], guard.StatusPassed)
	}

	if result["passed"] != true {
		t.Fatalf("result.passed = %#v, want true", result["passed"])
	}

	if _, ok := result["FixedData"]; ok {
		t.Fatalf("FixedData is present, want omitted: %s", raw)
	}

	if _, ok := result["fixed_data"]; ok {
		t.Fatalf("fixed_data is present, want omitted: %s", raw)
	}
}

func TestEncodeEnvelope_ResultWithValidationError(t *testing.T) {
	raw := encodeEnvelope(wasmValidateEnvelope{
		OK:    true,
		Error: "validation exploded",
		Result: &guard.ValidateResponse{
			Status:  guard.StatusFailed,
			Failed:  true,
			Errored: true,
			Error:   "validation exploded",
		},
	})

	var env wasmValidateEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("json.Unmarshal failed: %v\nraw: %s", err, raw)
	}

	if !env.OK {
		t.Fatal("OK = false, want true")
	}

	if env.Error != "validation exploded" {
		t.Fatalf("Error = %q, want %q", env.Error, "validation exploded")
	}

	if env.Result == nil {
		t.Fatal("Result = nil, want result")
	}

	if env.Result.Status != guard.StatusFailed {
		t.Fatalf("Result.Status = %q, want %q", env.Result.Status, guard.StatusFailed)
	}
}
