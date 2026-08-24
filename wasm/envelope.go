//go:build js && wasm

package main

import (
	"encoding/json/v2"
	"fmt"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

type wasmValidateEnvelope struct {
	OK     bool                    `json:"ok"`
	Result *guard.ValidateResponse `json:"result,omitzero"`
	Error  string                  `json:"error,omitzero"`
}

func encodeEnvelope(resp wasmValidateEnvelope) string {
	data, err := json.Marshal(resp)
	if err == nil {
		return string(data)
	}

	fallback, _ := json.Marshal(
		errorEnvelope(
			fmt.Sprintf("failed to encode response: %v", err),
		),
	)

	return string(fallback)
}

func errorEnvelope(msg string) wasmValidateEnvelope {
	return wasmValidateEnvelope{
		OK:    false,
		Error: msg,
	}
}
