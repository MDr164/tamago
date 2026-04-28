// ASPEED AST2600 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2600

// uartLcrWordLen8 sets 8-bit word length in LCR (bits[1:0] = 0b11).
const uartLcrWordLen8 = 0x3

// uartLcrDLAB sets the Divisor Latch Access Bit in LCR (bit 7).
const uartLcrDLAB = 1 << 7

// uartFifoEnable sets FCR bit 0 to enable FIFO mode.
const uartFifoEnable = 0x1

// uartLsrTxHoldEmpty is LSR bit 5: transmitter holding register empty.
const uartLsrTxHoldEmpty = 1 << 5

// uartBaudDiv115200 is the 16550 divisor for 115200 baud with the AST2600
// default UART clock of 24 MHz / 13 ≈ 1,846,153 Hz:
//   divisor = 1,846,153 / (16 * 115200) = 1.0 → DLL=1, DLH=0
const uartBaudDiv115200 = 1

// Init configures UART5 for 115200-8N1 operation.
func (u *Uart) Init() {
	// Set DLAB=1 to access baud rate divisor registers at offsets 0x00/0x04.
	u.WriteLCR(uartLcrDLAB)
	// DLL=1, DLH=0 → 115200 baud at ~1.846 MHz UART clock.
	u.WriteTHR(uartBaudDiv115200) // DLL at offset 0x00 (DLAB=1)
	u.WriteIER(0)                 // DLH at offset 0x04 (DLAB=1)
	// Set DLAB=0, configure 8 data bits, no parity, 1 stop bit.
	u.WriteLCR(uartLcrWordLen8)
	// Enable and reset FIFOs.
	u.WriteFCR(uartFifoEnable | 0x6)
}

// Tx transmits a single byte, blocking until the transmit buffer is empty.
func (u *Uart) Tx(c byte) {
	for u.ReadLSR()&uartLsrTxHoldEmpty == 0 {
	}
	u.WriteTHR(uint32(c))
}

// Write transmits a byte slice.
func (u *Uart) Write(b []byte) {
	for _, c := range b {
		u.Tx(c)
	}
}
