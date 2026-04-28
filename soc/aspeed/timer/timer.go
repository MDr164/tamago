// ASPEED FTTMR010 timer driver for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package timer implements a driver for the ASPEED FTTMR010-derived timer
// controller present on AST2400, AST2500, and AST2600 SoCs at 0x1E782000.
//
// The controller provides 8 independent 32-bit decrement counters, each
// selectable between PCLK and a fixed 1 MHz clock.
//
// Register layout (base = 0x1E782000):
//   Timer N base offset: N=1→0x00, N=2→0x10, N=3→0x20, N=4→0x40, …
//   +0x00  TMCx0  Counter status (current value; also write initial value)
//   +0x04  TMCx4  Reload value (loaded on counter reaching 0)
//   +0x08  TMCx8  Match 1 register (set 0xFFFFFFFF to disable)
//   +0x0C  TMCxC  Match 2 register (set 0xFFFFFFFF to disable)
//   +0x30  TMC30  Control register (4 bits per timer; W1S)
//   +0x34  TMC34  Interrupt status (bit N-1 for Timer N; W1C)
//   +0x3C  TMC3C  Control clear register (W1C; mirrors TMC30)
//
// TMC30 bit layout per timer (Timer 1 at bits [3:0], Timer 2 at [7:4], …):
//   bit+0  enable         1 = timer running
//   bit+1  1MHz clock     1 = fixed 1 MHz; 0 = PCLK
//   bit+2  overflow IRQ   1 = generate interrupt on counter wrap
//   bit+3  WDT reset      1 = trigger watchdog reset on overflow
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package timer

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// Timer register offsets (from controller base).
const (
	// Timer 1 (N=1)
	t1Status = 0x00
	t1Reload = 0x04
	t1Match1 = 0x08
	t1Match2 = 0x0C

	// Timer 2 (N=2)
	t2Status = 0x10
	t2Reload = 0x14
	t2Match1 = 0x18
	t2Match2 = 0x1C

	// Shared control
	ctrl    = 0x30
	irqStat = 0x34
	ctrlClr = 0x3C
)

// TMC30 control bits.
const (
	T1En     = 1 << 0 // Timer 1 enable
	T1_1MHz  = 1 << 1 // Timer 1 use 1 MHz clock
	T1OvfIRQ = 1 << 2 // Timer 1 overflow interrupt enable
	T2En     = 1 << 4 // Timer 2 enable
	T2_1MHz  = 1 << 5 // Timer 2 use 1 MHz clock
	T2OvfIRQ = 1 << 6 // Timer 2 overflow interrupt enable
)

// T2IRQBit is the interrupt status bit for Timer 2 in TMC34.
const T2IRQBit = 1 << 1

// Timer represents an ASPEED FTTMR010 timer controller instance.
type Timer struct {
	// Base address of the timer controller (e.g. 0x1E782000)
	Base uint32

	// Unexported nanotime state — software-extended 64-bit counter.
	last uint32
	high uint64
}

// InitFreeRunning configures Timer 1 as a 1 MHz free-running countdown
// counter for use as the nanotime source. Timer 1 counts from 0xFFFFFFFF
// downward and reloads automatically. No interrupt is generated.
func (t *Timer) InitFreeRunning() {
	// Stop and reset Timer 1.
	reg.Write(t.Base+ctrlClr, T1En|T1OvfIRQ)
	// Set reload value to maximum (24-bit or 32-bit depending on SoC).
	reg.Write(t.Base+t1Reload, 0xFFFFFFFF)
	// Disable match interrupts.
	reg.Write(t.Base+t1Match1, 0xFFFFFFFF)
	reg.Write(t.Base+t1Match2, 0xFFFFFFFF)
	// Set initial counter value.
	reg.Write(t.Base+t1Status, 0xFFFFFFFF)
	// Reset 64-bit extension state.
	t.last = 0xFFFFFFFF
	t.high = 0
	// Enable Timer 1 with 1 MHz clock.
	reg.Write(t.Base+ctrl, T1En|T1_1MHz)
}

// InitPeriodic configures Timer 2 as a periodic 1 MHz countdown that
// generates an overflow interrupt every periodUs microseconds.
// The interrupt must be acknowledged with AckT2IRQ after each firing.
func (t *Timer) InitPeriodic(periodUs uint32) {
	// Stop Timer 2.
	reg.Write(t.Base+ctrlClr, T2En|T2OvfIRQ)
	// Set reload = period (counts down from periodUs to 0 at 1 MHz → fires at T µs).
	reg.Write(t.Base+t2Reload, periodUs-1)
	// Disable match interrupts.
	reg.Write(t.Base+t2Match1, 0xFFFFFFFF)
	reg.Write(t.Base+t2Match2, 0xFFFFFFFF)
	// Set initial counter value.
	reg.Write(t.Base+t2Status, periodUs-1)
	// Enable Timer 2 with 1 MHz clock and overflow interrupt.
	reg.Write(t.Base+ctrl, T2En|T2_1MHz|T2OvfIRQ)
}

// Nanotime returns monotonic time in nanoseconds derived from Timer 1.
// Timer 1 counts DOWN at 1 MHz; each tick = 1 µs = 1000 ns.
// The 32-bit counter is software-extended to 64 bits by detecting wraps
// (a new value larger than the last value indicates wrap-around).
func (t *Timer) Nanotime() int64 {
	now := reg.Read(t.Base + t1Status)
	if now > t.last {
		t.high += uint64(0xFFFFFFFF) + 1
	}
	t.last = now
	elapsed := t.high + uint64(0xFFFFFFFF-now)
	return int64(elapsed * 1000)
}

// ReadT1Status returns the current Timer 1 countdown value.
// Used by the nanotime implementation for 64-bit time extension.
func (t *Timer) ReadT1Status() uint32 {
	return reg.Read(t.Base + t1Status)
}

// AckT2IRQ clears the Timer 2 overflow interrupt status bit (TMC34 bit 1).
// Must be called from the interrupt service routine after each Timer 2 tick.
func (t *Timer) AckT2IRQ() {
	reg.Write(t.Base+irqStat, T2IRQBit)
}
