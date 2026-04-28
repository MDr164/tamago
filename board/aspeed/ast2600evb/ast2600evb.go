// ASPEED AST2600EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ast2600evb provides hardware initialization for the ASPEED
// AST2600 Evaluation Board (dual-core Cortex-A7, 512 MB DDR4).
//
// Build with:
//
//	GOOS=tamago GOARCH=arm GOARM=7,softfloat \
//	  go tool tamago build \
//	    -tags linkcpuinit,softfloat \
//	    -ldflags "-T 0x80010000 -R 0x1000" \
//	    ./cmd/myapp
//
// To enable interrupt-driven scheduler ticks the application should call
// arm.ServiceInterrupts with an ISR once after Init returns, for example:
//
//	arm.ServiceInterrupts(func() {
//	    id := ast2600evb.GIC.GetInterrupt()
//	    if id == ast2600.TIMER2_IRQ {
//	        ast2600.AckTimerIRQ()
//	    }
//	    ast2600evb.GIC.EndInterrupt(id) // no-op, EOI is in GetInterrupt
//	})
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package ast2600evb

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aspeed/ast2600"
)

// Peripheral instances re-exported for application convenience.
var (
	UART5 = ast2600.UART5
	GIC   = ast2600.GIC
	ARM   = ast2600.ARM
)

// Init performs board-level hardware initialization triggered early in
// runtime setup (post Go-World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	ast2600.Init()
	ast2600.UART5.Init()
	ast2600.GIC.EnableInterrupt(ast2600.TIMER2_IRQ)
}
