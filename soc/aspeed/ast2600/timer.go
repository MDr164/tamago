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

// TimerCtrl bit positions (4 bits per timer, Timer N at bit 4*(N-1)).
const (
	timerCtrlT1En     = 0  // Timer 1 enable
	timerCtrlT1_1MHz  = 1  // Timer 1 use 1 MHz clock
	timerCtrlT1OvfIRQ = 2  // Timer 1 overflow interrupt enable
	timerCtrlT2En     = 4  // Timer 2 enable
	timerCtrlT2_1MHz  = 5  // Timer 2 use 1 MHz clock
	timerCtrlT2OvfIRQ = 6  // Timer 2 overflow interrupt enable
)

// timerLast is the last raw 32-bit counter value read.
var timerLast uint32

// timerHigh tracks upper bits for 64-bit software-extended nanotime.
// Timer 1 counts DOWN from 0xFFFFFFFF at 1 MHz, so each tick = 1 µs.
var timerHigh uint64

// initTimers configures:
//   - Timer 1: free-running 1 MHz countdown counter for nanotime
//   - Timer 2: periodic overflow interrupt at timerTickPeriodUs for scheduler tick
//
// The ARM generic timer (CNTFRQ/CNTPCT via CP15) is available on Cortex-A7
// but requires CNTFRQ to be set in EL1 before use. We use the ASPEED MMIO timer
// instead to avoid the CNTFRQ dependency at early boot.
//
// Note: If ARM.InitGenericTimers is preferred, replace this with a call to
// ARM.InitGenericTimers(0, CPUFreqHz()) in Init() and remove this file.
func initTimers() {
	// Disable both timers before configuring.
	TIMER.WriteCTRLCLR((1 << timerCtrlT1En) | (1 << timerCtrlT2En))

	// Timer 1: free-running 1 MHz countdown, no interrupt, reload from max.
	TIMER.WriteT1RELOAD(0xFFFFFFFF)
	TIMER.WriteT1MATCH1(0xFFFFFFFF) // disable match interrupt
	TIMER.WriteT1MATCH2(0xFFFFFFFF)
	TIMER.WriteT1STATUS(0xFFFFFFFF) // set initial value
	timerLast = 0xFFFFFFFF
	timerHigh = 0

	// Timer 2: 1 MHz, overflow interrupt every timerTickPeriodUs microseconds.
	const timerTickPeriodUs = 10000 // 10 ms
	TIMER.WriteT2RELOAD(timerTickPeriodUs - 1)
	TIMER.WriteT2MATCH1(0xFFFFFFFF)
	TIMER.WriteT2MATCH2(0xFFFFFFFF)
	TIMER.WriteT2STATUS(timerTickPeriodUs - 1)

	// Enable both timers with 1 MHz clock; Timer 2 with overflow interrupt.
	TIMER.WriteCTRL((1 << timerCtrlT1En) | (1 << timerCtrlT1_1MHz) |
		(1 << timerCtrlT2En) | (1 << timerCtrlT2_1MHz) | (1 << timerCtrlT2OvfIRQ))
}

// AckTimerIRQ clears the Timer 2 overflow interrupt status bit (bit 1 of TMC34).
// Called from the interrupt service routine.
func AckTimerIRQ() {
	TIMER.WriteIRQSTATUS(1 << 1)
}

// readTimer reads the current Timer 1 countdown value.
func readTimer() uint32 {
	return TIMER.ReadT1STATUS()
}

// nanotime returns monotonic time in nanoseconds, derived from Timer 1.
// Timer 1 counts DOWN at 1 MHz, so each tick = 1 µs = 1000 ns.
// We extend the 32-bit counter to 64 bits in software by tracking wrap-arounds.
//
//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	now := readTimer()
	// Timer counts DOWN. A new value greater than last means a wrap-around.
	if now > timerLast {
		timerHigh += uint64(0xFFFFFFFF) + 1
	}
	timerLast = now
	// Convert µs (1 MHz ticks, counting down from 0xFFFFFFFF) to ns.
	// Elapsed ticks = 0xFFFFFFFF - now + timerHigh_offset
	elapsed := timerHigh + uint64(0xFFFFFFFF-now)
	return int64(elapsed * 1000)
}
