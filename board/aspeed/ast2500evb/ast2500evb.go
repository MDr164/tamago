// ASPEED AST2500EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ast2500evb provides hardware initialization for the ASPEED
// AST2500 Evaluation Board (ARM1176JZS ARMv6, 512 MB DDR3).
//
// Build with:
//
//	GOOS=tamago GOARCH=arm GOARM=6,softfloat \
//	  go tool tamago build \
//	    -tags linkcpuinit,softfloat \
//	    -ldflags "-T 0x80010000 -R 0x1000" \
//	    ./cmd/myapp
//
// The linker text base is 0x80010000 (canonical SDRAM) so QEMU can load the
// ELF into physical SDRAM. cpuinit.s then sets AHBC8C[0]=1 to remap SDRAM
// to 0x00000000; goos.RamStart=0x0 so exception vectors and the Go heap use
// the 0x0-aliased SDRAM. Both 0x0 and 0x80000000 address the same memory.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package ast2500evb

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aspeed/ast2500"
)

// Peripheral instances re-exported for application convenience.
var (
	UART5 = ast2500.UART5
	VIC   = ast2500.VIC
	ARM   = ast2500.ARM
)

// Init performs board-level hardware initialization triggered early in
// runtime setup (post Go-World start).
//
// Applications must call arm.ServiceInterrupts with an ISR function to enable
// interrupt delivery for the scheduler timer tick:
//
//	arm.ServiceInterrupts(func() {
//	    if id := ast2500.VIC.CurrentIRQ(); id == ast2500.TIMER2_IRQ {
//	        ast2500.AckTimerIRQ()
//	    }
//	})
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	ast2500.Init()
	ast2500.UART5.Init()
	// Timer2 IRQ is configured by initTimers() but not enabled in the VIC here.
	// Enable it from application code after arm.ServiceInterrupts is running
	// and exception vector handling at 0x0 is confirmed working on the target.
	// ast2500.VIC.EnableIRQ(ast2500.TIMER2_IRQ)
}
