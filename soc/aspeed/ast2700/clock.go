// ASPEED AST2700 SoC support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2700

import "github.com/usbarmory/tamago/internal/reg"

// AST2700 SCU protection key (same as AST2600)
const scuProtKey = 0x1688a8a8

// SCU0 (CPU die) register offsets
const (
	scuProtReg  = 0x000 // Protection Key Register
	scuRevID    = 0x004 // Silicon Revision ID
	scuClkStop1 = 0x240 // Clock Stop Control 1
	scuClkClr1  = 0x244 // Clock Stop Clear 1
)

// SCUIO (IO die) register offsets
const (
	scuiRevID = 0x000 // Silicon Revision ID
)

// scuUnlock enables write access to SCU registers.
func scuUnlock() {
	reg.Write(SCU_BASE+scuProtReg, scuProtKey)
}

// scuLock re-enables SCU write protection.
func scuLock() {
	reg.Write(SCU_BASE+scuProtReg, 0x1)
}

// EnableUART12Clock ensures the UART12 clock is running.
func EnableUART12Clock() {
	scuUnlock()
	// Write 1 to clear the clock-stop bit for UART12
	reg.Write(SCU_BASE+scuClkClr1, 1<<15)
	scuLock()
}

// CPUFreqHz returns the Cortex-A35 CPU clock frequency in Hz.
// On QEMU the CPU frequency is 1125000000 Hz (1.125 GHz).
func CPUFreqHz() uint32 {
	return 1_125_000_000
}
