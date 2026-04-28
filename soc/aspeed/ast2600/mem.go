// ASPEED AST2600 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package ast2600

import _ "unsafe"

// SDRAM is mapped at 0x80000000 on AST2600. The ASPEED boot ROM initializes
// SDRAM and loads the TamaGo image there before jumping to the entry point.
//
//go:linkname ramStart runtime/goos.RamStart
var ramStart uint32 = 0x80000000

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB
