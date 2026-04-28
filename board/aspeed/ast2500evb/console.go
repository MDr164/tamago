// ASPEED AST2500EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package ast2500evb

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/aspeed/ast2500"
)

// printk routes single-byte runtime console output to UART5.
// Writes directly to THR without checking LSR so that early panics
// (before UART5.Init()) still produce output.
//
//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	if c == '\n' {
		reg.Write(ast2500.UART5_BASE, uint32('\r'))
	}
	reg.Write(ast2500.UART5_BASE, uint32(c))
}
