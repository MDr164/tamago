// ASPEED AST2700 DCSCM board support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkprintk

package ast2700dcscm

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aspeed/ast2700"
)

// printk routes single-byte runtime console output to UART12.
// LSR polling ensures characters are not dropped on real hardware.
//
//go:linkname printk runtime/goos.Printk
func printk(c byte) {
	if c == '\n' {
		ast2700.UART12.Tx('\r')
	}
	ast2700.UART12.Tx(c)
}
