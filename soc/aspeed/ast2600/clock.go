// ASPEED AST2600 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2600

// scuUnlock enables write access to SCU registers.
// The AST2600 SCU requires writing the magic value 0x1688A8A8 to SCU000
// before any SCU register can be modified.
func scuUnlock() {
	SCU.WritePROTECTIONKEY(0x1688A8A8)
}

// scuLock re-enables SCU write protection.
func scuLock() {
	SCU.WritePROTECTIONKEY(0x1)
}

// EnableUART5Clock ensures the UART5 clock is running.
// SCU080 bit 15 = UART5CLK stop (1=stopped). Writing 1 to SCU084 bit 15 clears
// the stop, i.e. starts the clock.
func EnableUART5Clock() {
	scuUnlock()
	SCU.WriteCLKSTOPCLR1(1 << 15)
	scuLock()
}

// EnableHACEClock ensures the Hash and Crypto Engine (HACE/RNG) clock is running.
// SCU080 bit 13 = HACE/RNG clock stop. Writing 1 to SCU084 bit 13 starts it.
func EnableHACEClock() {
	scuUnlock()
	SCU.WriteCLKSTOPCLR1(1 << 13)
	scuLock()
}

// CPUFreqHz returns the Cortex-A7 CPU clock frequency in Hz, derived from
// HW_STRAP1 (SCU500) bits [10:8].
//
// The frequency is used to program the ARM generic timer CNTFRQ register.
// (AST2600 datasheet, SCU500 register description)
func CPUFreqHz() uint32 {
	strap := SCU.ReadHWSTRAP1()
	sel := (strap >> 8) & 0x7
	switch sel {
	case 0, 2:
		return 1_200_000_000 // 1.2 GHz
	case 1, 3:
		return 1_600_000_000 // 1.6 GHz
	default:
		return 800_000_000 // 800 MHz
	}
}
