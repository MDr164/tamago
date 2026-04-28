// ASPEED VIC interrupt controller driver for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package intc implements a driver for the ASPEED VIC (Vectored Interrupt
// Controller) present on AST2400 and AST2500 SoCs.
//
// Physical base address: 0x1E6C0000
// IRQ count: 51 (AST2400), 64 (AST2500)
//
// The VIC provides two register blocks:
//   Legacy (base+0x00): covers IRQ0–31 only
//   New Mapping (base+0x80): covers IRQ0–31 (low) and IRQ32–63 (high)
//
// This driver uses the Legacy block for Timer IRQs (IRQ 16 and 17) and the
// New Mapping high registers for IRQs 32–63 on AST2500.
//
// Note: AST2600 uses an ARM GIC instead (arm/gic package).
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package intc

import (
	"github.com/usbarmory/tamago/internal/reg"
)

// VIC register offsets from base address (0x1E6C0000).
// Source: AST2400 datasheet §VIC, AST2500 datasheet §VIC.
const (
	// Legacy mapping — IRQ0–31 only.
	irqStatus    = 0x00 // IRQ Status (active + enabled)
	fiqStatus    = 0x04 // FIQ Status
	rawStatus    = 0x08 // Raw Interrupt Status (pre-mask)
	intSelect    = 0x0C // Interrupt Selection (0=IRQ, 1=FIQ)
	intEnable    = 0x10 // Interrupt Enable (W1S — write 1 to enable)
	intEnClear   = 0x14 // Interrupt Enable Clear (W1C — write 1 to disable)
	softInt      = 0x18 // Software Interrupt
	softIntClear = 0x1C // Software Interrupt Clear
	edgeIntClear = 0x38 // Edge-triggered Interrupt Clear

	// New Mapping — Low (IRQ0–31) and High (IRQ32–63) word pairs.
	// Used for AST2500 where IRQ 32–63 may be needed.
	newIRQStatusL  = 0x80
	newIRQStatusH  = 0x84
	newIntEnableL  = 0xA0
	newIntEnableH  = 0xA4
	newIntEnClearL = 0xA8
	newIntEnClearH = 0xAC
)

// INTC represents an ASPEED VIC instance.
type INTC struct {
	// Base address of the VIC (0x1E6C0000).
	Base uint32
}

// Init disables all interrupt sources. Call once before enabling specific IRQs.
func (hw *INTC) Init() {
	// Disable all IRQ0–31.
	reg.Write(hw.Base+intEnClear, 0xFFFFFFFF)
	// Disable all IRQ32–63 (AST2500 only; write is safe on AST2400 too).
	reg.Write(hw.Base+newIntEnClearH, 0xFFFFFFFF)
	// Ensure all interrupts route to IRQ (not FIQ).
	reg.Write(hw.Base+intSelect, 0)
}

// EnableIRQ enables forwarding of interrupt source n to the CPU.
// Valid range: 0–31 (Legacy registers) or 32–63 (New Mapping High registers).
func (hw *INTC) EnableIRQ(n int) {
	if n < 32 {
		reg.Write(hw.Base+intEnable, 1<<uint(n))
	} else {
		reg.Write(hw.Base+newIntEnableH, 1<<uint(n-32))
	}
}

// DisableIRQ disables forwarding of interrupt source n.
func (hw *INTC) DisableIRQ(n int) {
	if n < 32 {
		reg.Write(hw.Base+intEnClear, 1<<uint(n))
	} else {
		reg.Write(hw.Base+newIntEnClearH, 1<<uint(n-32))
	}
}

// CurrentIRQ returns the lowest-numbered active, enabled interrupt, or -1 if
// no interrupt is pending. Called from the ARM IRQ exception handler (ISR).
func (hw *INTC) CurrentIRQ() int {
	// Check legacy IRQ0–31.
	status := reg.Read(hw.Base + irqStatus)
	for i := 0; i < 32; i++ {
		if status&(1<<uint(i)) != 0 {
			return i
		}
	}
	// Check high IRQ32–63 (New Mapping).
	statusH := reg.Read(hw.Base + newIRQStatusH)
	for i := 0; i < 32; i++ {
		if statusH&(1<<uint(i)) != 0 {
			return i + 32
		}
	}
	return -1
}
