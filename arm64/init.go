// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
)

// debugUART12THR is the AST2700 UART12 transmit-holding register. Used only for
// raw boot breadcrumbs during CA35 bring-up (32-bit-spaced 16550 registers, so
// a 32-bit write hits THR alone). DEBUG — remove once bring-up is stable.
const debugUART12THR = 0x14c33b00

// Init takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
//go:linkname Init runtime/goos.Hwinit0
func Init() {
	// DEBUG breadcrumb 'e': entered Hwinit0 (MMU still off).
	reg.Write(debugUART12THR, 'e')
	fp_enable()
	var cpu CPU
	cpu.InitMMU()
	// DEBUG breadcrumb 'f': InitMMU returned — MMU + I/D caches now enabled and
	// we survived executing post-enable instructions.
	reg.Write(debugUART12THR, 'f')
}
