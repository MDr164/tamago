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
)

// FTTMR010 TMC30 control register bit positions (4 bits per timer).
const (
	tmcT1En     = 1 << 0 // Timer 1 enable
	tmcT1_1MHz  = 1 << 1 // Timer 1 use 1 MHz clock
	tmcT1OvfIRQ = 1 << 2 // Timer 1 overflow interrupt enable
	tmcT2En     = 1 << 4 // Timer 2 enable
	tmcT2_1MHz  = 1 << 5 // Timer 2 use 1 MHz clock
	tmcT2OvfIRQ = 1 << 6 // Timer 2 overflow interrupt enable
)

// T2IRQBit is the interrupt status bit for Timer 2 in TMC34 (bit 1).
const T2IRQBit = 1 << 1

// timerTickPeriodUs is the Timer 2 interrupt period in microseconds (10 ms).
const timerTickPeriodUs = 10000

// nanotimeLast and nanotimeHigh provide software 64-bit extension of Timer 1.
// Must be package-level variables since nanotime is called by the Go runtime
// before package init() functions run.
var nanotimeLast uint32
var nanotimeHigh uint64

// initTimers configures:
//   - Timer 1: free-running 1 MHz countdown counter for nanotime
//   - Timer 2: periodic overflow interrupt at timerTickPeriodUs for scheduler tick
func initTimers() {
	// Stop timers.
	TIMER.WriteCTRLCLR(tmcT1En | tmcT1OvfIRQ | tmcT2En | tmcT2OvfIRQ)

	// Timer 1: free-running 1 MHz countdown from 0xFFFFFFFF.
	TIMER.WriteT1RELOAD(0xFFFFFFFF)
	TIMER.WriteT1MATCH1(0xFFFFFFFF)
	TIMER.WriteT1MATCH2(0xFFFFFFFF)
	TIMER.WriteT1STATUS(0xFFFFFFFF)
	nanotimeLast = 0xFFFFFFFF
	nanotimeHigh = 0

	// Timer 2: 1 MHz periodic with overflow interrupt every timerTickPeriodUs µs.
	TIMER.WriteT2RELOAD(timerTickPeriodUs - 1)
	TIMER.WriteT2MATCH1(0xFFFFFFFF)
	TIMER.WriteT2MATCH2(0xFFFFFFFF)
	TIMER.WriteT2STATUS(timerTickPeriodUs - 1)

	// Enable both timers.
	TIMER.WriteCTRL((tmcT1En | tmcT1_1MHz) | (tmcT2En | tmcT2_1MHz | tmcT2OvfIRQ))
}

// AckTimerIRQ clears the Timer 2 overflow interrupt status bit (TMC34 bit 1).
// Must be called from the interrupt service routine.
func AckTimerIRQ() {
	TIMER.WriteIRQSTATUS(T2IRQBit)
}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	// Timer 1 counts DOWN at 1 MHz; each tick = 1 µs = 1000 ns.
	now := TIMER.ReadT1STATUS()
	if now > nanotimeLast {
		nanotimeHigh += uint64(0xFFFFFFFF) + 1
	}
	nanotimeLast = now
	elapsed := nanotimeHigh + uint64(0xFFFFFFFF-now)
	return int64(elapsed * 1000)
}
