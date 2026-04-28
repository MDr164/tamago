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
)

// FTTMR010 TMC30 control register bit positions.
const (
	tmcT1En     = 1 << 0
	tmcT1_1MHz  = 1 << 1
	tmcT1OvfIRQ = 1 << 2
	tmcT2En     = 1 << 4
	tmcT2_1MHz  = 1 << 5
	tmcT2OvfIRQ = 1 << 6
)

// T2IRQBit is the interrupt status bit for Timer 2 in TMC34 (bit 1).
const T2IRQBit = 1 << 1

// timerTickPeriodUs is the Timer 2 interrupt period in microseconds (10 ms).
const timerTickPeriodUs = 10000

// nanotimeLast and nanotimeHigh provide software 64-bit extension of Timer 1.
// Must be package-level since nanotime is called before package init().
var nanotimeLast uint32
var nanotimeHigh uint64

// EnableIdleWFI sets up a no-op idle governor initially.
// For ARM1176 with the VIC-based timer tick, call this after arm.ServiceInterrupts
// is running to switch to a power-saving idle. Currently left as no-op since
// ARM1176 WFI uses CP15 (already handled by arm/irq.s for arm.6 builds).
func EnableIdleWFI() {}

// initTimers configures:
//   - Timer 1: free-running 1 MHz countdown counter for nanotime
//   - Timer 2: periodic overflow interrupt at timerTickPeriodUs for scheduler tick
func initTimers() {
	TIMER.WriteCTRLCLR(tmcT1En | tmcT1OvfIRQ | tmcT2En | tmcT2OvfIRQ)

	TIMER.WriteT1RELOAD(0xFFFFFFFF)
	TIMER.WriteT1MATCH1(0xFFFFFFFF)
	TIMER.WriteT1MATCH2(0xFFFFFFFF)
	TIMER.WriteT1STATUS(0xFFFFFFFF)
	nanotimeLast = 0xFFFFFFFF
	nanotimeHigh = 0

	TIMER.WriteT2RELOAD(timerTickPeriodUs - 1)
	TIMER.WriteT2MATCH1(0xFFFFFFFF)
	TIMER.WriteT2MATCH2(0xFFFFFFFF)
	TIMER.WriteT2STATUS(timerTickPeriodUs - 1)

	TIMER.WriteCTRL((tmcT1En | tmcT1_1MHz) | (tmcT2En | tmcT2_1MHz | tmcT2OvfIRQ))
}

// AckTimerIRQ clears the Timer 2 overflow interrupt.
func AckTimerIRQ() {
	TIMER.WriteIRQSTATUS(T2IRQBit)
}

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	now := TIMER.ReadT1STATUS()
	if now > nanotimeLast {
		nanotimeHigh += uint64(0xFFFFFFFF) + 1
	}
	nanotimeLast = now
	elapsed := nanotimeHigh + uint64(0xFFFFFFFF-now)
	return int64(elapsed * 1000)
}
