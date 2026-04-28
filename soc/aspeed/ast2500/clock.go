// ASPEED AST2500 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2500

// scuUnlock enables write access to SCU registers.
func scuUnlock() {
	SCU.WritePROTECTIONKEY(0x1688A8A8)
}

// scuLock re-enables SCU write protection.
func scuLock() {
	SCU.WritePROTECTIONKEY(0x1)
}

// EnableUART5Clock ensures the UART5 clock is running.
// SCU0C bit 15 = UART5 clock stop; writing 0 to it enables the clock.
func EnableUART5Clock() {
	scuUnlock()
	// Clear bit 15 (UART5 clock stop) in SCU0C.
	v := SCU.ReadCLKSTOP1()
	v &^= 1 << 15
	SCU.WriteCLKSTOP1(v)
	scuLock()
}
