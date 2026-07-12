// ASPEED AST2700 SoC support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2700

// Peripheral base addresses (AST2700 silicon memory map)
const (
	// SCU_BASE System Control Unit (CPU die)
	SCU_BASE = 0x12c02000
	// SCUIO_BASE System Control Unit (IO die)
	SCUIO_BASE = 0x14c02000
	// UART12_BASE UART12 (BMC debug console, 16550-compatible)
	UART12_BASE = 0x14c33b00
	// TIMER_BASE Timer Controller
	TIMER_BASE = 0x12c10000
	// GIC_DIST_BASE GICv3 Distributor
	GIC_DIST_BASE = 0x12200000
	// GIC_REDIST_BASE GICv3 Redistributor
	GIC_REDIST_BASE = 0x12280000
	// TRNG_BASE True Random Number Generator
	TRNG_BASE = 0x14c3b000
)

// IRQ numbers (GIC SPI numbers, offset by -32 from datasheet GICINT)
const (
	// TIMER1_IRQ Timer 1 overflow (GIC SPI 16)
	TIMER1_IRQ = 16
	// TIMER2_IRQ Timer 2 overflow (GIC SPI 17)
	TIMER2_IRQ = 17
)
