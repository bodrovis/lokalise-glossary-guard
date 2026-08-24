//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"syscall/js"

	"github.com/bodrovis/lokalise-glossary-guard/pkg/guard"
)

func validateGlossaryGuard(_ js.Value, args []js.Value) (out any) {
	defer func() {
		if r := recover(); r != nil {
			out = encodeEnvelope(
				errorEnvelope(fmt.Sprintf("wasm panic: %v", r)),
			)
		}
	}()

	return encodeEnvelope(
		validateGlossaryGuardEnvelope(args),
	)
}

func validateGlossaryGuardEnvelope(args []js.Value) wasmValidateEnvelope {
	req, err := validateRequestFromArgs(args)
	if err != nil {
		return errorEnvelope(err.Error())
	}

	resp, validationErr := guard.ValidateBytes(
		context.Background(),
		req,
	)

	envelope := wasmValidateEnvelope{
		OK:     true,
		Result: &resp,
	}

	// Validation errors are returned as part of the structured result,
	// not treated as WASM invocation failures.
	if validationErr != nil {
		envelope.Error = validationErr.Error()
	}

	return envelope
}
