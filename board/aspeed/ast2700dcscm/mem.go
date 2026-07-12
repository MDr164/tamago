// ASPEED AST2700 DCSCM board support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package ast2700dcscm

import _ "unsafe"

// The DCSCM DDR4 board has 1 GB of physical DRAM.
//
//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x40000000
