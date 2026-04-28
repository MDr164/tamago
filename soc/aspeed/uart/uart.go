// ASPEED UART driver for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package uart implements a 16550-compatible UART driver for ASPEED AST-series
// SoCs. All ASPEED BMC chips (AST2400, AST2500, AST2600) use the same 16550
// UART IP with identical register layout.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package uart

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// 16550 UART register offsets.
const (
	rbrThr = 0x00 // Receive Buffer / Transmitter Holding Register (DLAB=0)
	ier    = 0x04 // Interrupt Enable Register (DLAB=0)
	dll    = 0x00 // Divisor Latch Low (DLAB=1)
	dlh    = 0x04 // Divisor Latch High (DLAB=1)
	iirFcr = 0x08 // Interrupt Identity / FIFO Control Register
	lcr    = 0x0C // Line Control Register; bit7 = DLAB
	mcr    = 0x10 // Modem Control Register
	lsr    = 0x14 // Line Status Register
	msr    = 0x18 // Modem Status Register
)

// LSR register bits.
const (
	lsrRxReady    = 1 << 0 // Receive data ready
	lsrTxHoldEmpty = 1 << 5 // Transmitter holding register empty
	lsrTxEmpty    = 1 << 6 // Transmitter empty
)

// LCR register bits.
const (
	lcrWordLen8 = 0x3 // 8-bit word length (bits[1:0] = 0b11)
	lcrDLAB     = 1 << 7 // Divisor Latch Access Bit
)

// FCR control bits.
const (
	fcrEnable     = 0x1 // FIFO enable
	fcrRxReset    = 0x2 // Reset Rx FIFO
	fcrTxReset    = 0x4 // Reset Tx FIFO
)

// UART represents an ASPEED 16550-compatible UART instance.
type UART struct {
	// Base register address (e.g. 0x1E784000 for UART5)
	Base uint32

	// Index is the UART number (1-based), used for identification.
	Index int
}

// Init configures the UART for 115200-8N1 operation.
//
// The baud rate divisor assumes the AST2600 default UART clock of
// 24 MHz / 13 ≈ 1.846 MHz, giving divisor = 1 for 115200 baud.
// For AST2400/AST2500 the same UART clock source applies.
// Divisor formula: divisor = (refClock/13) / (16 * baud) = 1 for 115200.
func (u *UART) Init() {
	// Set DLAB=1 to program baud rate divisor registers at 0x00/0x04.
	reg.Write(u.Base+lcr, lcrDLAB)
	// Divisor = 1 → 115200 baud at ~1.846 MHz UART reference clock.
	reg.Write(u.Base+dll, 1) // DLL
	reg.Write(u.Base+dlh, 0) // DLH
	// Set DLAB=0, 8 data bits, no parity, 1 stop bit.
	reg.Write(u.Base+lcr, lcrWordLen8)
	// Enable and reset Rx/Tx FIFOs.
	reg.Write(u.Base+iirFcr, fcrEnable|fcrRxReset|fcrTxReset)
}

// Tx transmits a single byte. Blocks until the transmit holding register is
// empty (LSR bit 5). This is safe for real hardware after Init() is called.
//
// For early-boot output (before Init), use direct reg.Write to Base+0x00.
func (u *UART) Tx(c byte) {
	for reg.Read(u.Base+lsr)&lsrTxHoldEmpty == 0 {
	}
	reg.Write(u.Base+rbrThr, uint32(c))
}

// Write transmits a byte slice via Tx.
func (u *UART) Write(b []byte) {
	for _, c := range b {
		u.Tx(c)
	}
}
