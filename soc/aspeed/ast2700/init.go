// ASPEED AST2700 SoC support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2700

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/bits"
	"github.com/usbarmory/tamago/internal/reg"
)

// init runs once at package import. Reads the silicon revision from SCUIO.
func init() {
	SiliconRevision = HardwareRevision()
}

// RawRevision returns the SCUIO silicon revision register.
func RawRevision() uint32 {
	return reg.Read(SCUIO_BASE + scuiRevID)
}

// Generation returns the BMC generation field.
func Generation() uint32 {
	rev := RawRevision()
	return bits.GetN(&rev, 24, 0xFF)
}

// HardwareRevision returns the hardware revision field.
func HardwareRevision() uint32 {
	rev := RawRevision()
	return bits.GetN(&rev, 16, 0xFF)
}

// DeviceID returns the eFuse device ID field.
func DeviceID() uint32 {
	rev := RawRevision()
	return bits.GetN(&rev, 8, 0xFF)
}

// Model returns a human-readable SoC revision string.
func Model() string {
	switch DeviceID() {
	case 0x00:
		return "AST2750"
	case 0x01:
		return "AST2700"
	case 0x02:
		return "AST2720"
	default:
		return "AST27xx unknown"
	}
}

// Revision returns a human-readable silicon stepping string.
func Revision() string {
	switch HardwareRevision() {
	case 0x00:
		return "A0"
	case 0x01:
		return "A1"
	case 0x02:
		return "A2"
	default:
		return "unknown"
	}
}

// Init performs AST2700 SoC initialization. Must be called from Hwinit1
// (post Go-World start, memory allocation available).
func Init() {
	ARM.Init()
	ARM.InitGenericTimers(0, cpuFreqHz)
	ARM.EnableCache()

	GIC.Init()
	ARM.EnableInterrupts()

	EnableUART12Clock()
}
