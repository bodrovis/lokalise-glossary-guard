//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

type wasmValidateEnvelope struct {
	OK     bool                    `json:"ok"`
	Result *guard.ValidateResponse `json:"result,omitempty"`
	Error  string                  `json:"error,omitempty"`
}

func main() {
	js.Global().Set("validateGlossaryGuard", js.FuncOf(validateGlossaryGuard))

	// Keep WASM alive.
	select {}
}

func validateGlossaryGuard(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return encodeEnvelope(wasmValidateEnvelope{
			OK:    false,
			Error: "missing input",
		})
	}

	input, err := inputToJSON(args[0])
	if err != nil {
		return encodeEnvelope(wasmValidateEnvelope{
			OK:    false,
			Error: err.Error(),
		})
	}

	var req guard.ValidateRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		return encodeEnvelope(wasmValidateEnvelope{
			OK:    false,
			Error: fmt.Sprintf("invalid input json: %v", err),
		})
	}

	resp, validationErr := guard.ValidateBytes(context.Background(), req)

	envelope := wasmValidateEnvelope{
		OK:     true,
		Result: &resp,
	}

	// Validation failure is not a WASM crash.
	// The structured result is still returned.
	if validationErr != nil {
		envelope.Error = validationErr.Error()
	}

	return encodeEnvelope(envelope)
}

func inputToJSON(v js.Value) (string, error) {
	if v.Type() == js.TypeString {
		return v.String(), nil
	}

	jsonValue := js.Global().Get("JSON").Call("stringify", v)
	if jsonValue.IsUndefined() || jsonValue.IsNull() {
		return "", fmt.Errorf("input must be a JSON string or object")
	}

	return jsonValue.String(), nil
}

func encodeEnvelope(resp wasmValidateEnvelope) string {
	data, err := json.Marshal(resp)
	if err != nil {
		fallback, _ := json.Marshal(wasmValidateEnvelope{
			OK:    false,
			Error: fmt.Sprintf("failed to encode response: %v", err),
		})
		return string(fallback)
	}

	return string(data)
}
