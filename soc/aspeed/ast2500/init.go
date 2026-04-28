// ASPEED AST2500 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2500

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/bits"
)

// ScuRevIdHwRevPos is the bit position of the hardware revision in SCU7C.
const ScuRevIdHwRevPos = 16

func init() {
	rev := SCU.ReadREVID()
	SiliconRevision = bits.GetN(&rev, ScuRevIdHwRevPos, 0xFF)
}

// Model returns a human-readable SoC revision string.
func Model() string {
	switch SiliconRevision {
	case 0x00:
		return "AST2500 A0"
	case 0x01:
		return "AST2500 A1"
	case 0x02:
		return "AST2500 A2"
	default:
		return "AST2500 unknown"
	}
}

// Init performs AST2500 SoC initialization. Called from Hwinit1
// (post Go-World start, memory allocation available).
func Init() {
	if ARM.Mode() != arm.SYS_MODE {
		return
	}

	ARM.Init()
	// ARM1176 single-core: no EnableSMP.
	// MMU deferred: ARM.InitMMU() can be added when exception handling works.

	VIC.Init()
	// Note: ARM.EnableInterrupts() is NOT called here because the ARM hardware
	// exception vectors (fixed at 0x0 for NoVBAR=true) need to point to Go's
	// IRQ handlers before interrupts can be safely enabled. On real hardware
	// after AHBC remap, arm.Init() would write LDR PC table to 0x0 (remapped
	// SDRAM). On QEMU where AHBC remap is not implemented, enabling IRQs would
	// cause faults. Applications call arm.ServiceInterrupts() when ready.

	EnableUART5Clock()
	initTimers()
}
