// QEMU AST2700 support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package ast2700

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/aspeed/ast2700"
)

// printk routes single-byte runtime console output to UART12.
// Writes directly to THR without waiting for LSR TX-empty to ensure
// output works even before UART12.Init() is called.
//
//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	if c == '\n' {
		reg.Write(ast2700.UART12_BASE, uint32('\r'))
	}
	reg.Write(ast2700.UART12_BASE, uint32(c))
}
