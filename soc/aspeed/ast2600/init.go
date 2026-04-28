// ASPEED AST2600 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2600

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/bits"
)

// ScuRevIdHwRevPos is the bit position of the hardware revision field in SCU004.
const ScuRevIdHwRevPos = 16

// init runs once at package import. Reads the silicon revision from SCU004
// bits[23:16] and stores it in SiliconRevision.
func init() {
	rev := SCU.ReadREVID()
	SiliconRevision = bits.GetN(&rev, ScuRevIdHwRevPos, 0xFF)
}

// Model returns a human-readable SoC revision string.
func Model() string {
	switch SiliconRevision {
	case 0x00:
		return "AST2600 A0"
	case 0x01:
		return "AST2600 A1"
	case 0x02:
		return "AST2600 A2"
	case 0x03:
		return "AST2600 A3"
	default:
		return "AST2600 unknown"
	}
}

// Init performs AST2600 SoC initialization. Must be called from Hwinit1
// (post Go-World start, memory allocation available).
func Init() {
	if ARM.Mode() != arm.SYS_MODE {
		return
	}

	ARM.Init()
	ARM.EnableSMP()
	ARM.InitMMU()
	ARM.EnableCache()

	GIC.Init()
	ARM.EnableInterrupts(false)

	EnableUART5Clock()
	initTimers()
}
