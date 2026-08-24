//go:build js && wasm

package main

import "syscall/js"

func main() {
	validate := js.FuncOf(validateGlossaryGuard)

	js.Global().Set("validateGlossaryGuard", validate)

	// Keep WASM alive while the exported JS function is available.
	select {}
}
