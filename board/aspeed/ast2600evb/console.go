// ASPEED AST2600EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package ast2600evb

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/aspeed/ast2600"
)

// printk routes single-byte runtime console output to UART5.
// Writes directly to THR without waiting for LSR TX-empty to ensure
// output works even before UART5.Init() is called (e.g. in early panics).
// On QEMU the ASPEED UART accepts writes unconditionally; on real hardware
// the UART FIFO absorbs burst writes at boot time.
//
//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	if c == '\n' {
		reg.Write(ast2600.UART5_BASE, uint32('\r'))
	}
	reg.Write(ast2600.UART5_BASE, uint32(c))
}
