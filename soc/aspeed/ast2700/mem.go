// ASPEED AST2700 SoC support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package ast2700

import _ "unsafe"

// SDRAM is mapped at 0x400000000 on the AST2700 CA35 address space.
// The BootMCU initializes DRAM and loads the TamaGo image before release.
//
// RamStart is the base of DRAM as seen by the CA35. MMU page tables are
// placed at RamStart + 0x4000..0x7FFF. The diagnostic scratch area at
// RamStart + 0x3000 is reserved for BootMCU polling and is not managed
// by the Go runtime.
//
//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = 0x400000000

// RamStackOffset reserves space at the top of the runtime-managed region
// for the initial stack. SP = RamStart + RamSize - RamStackOffset.
//
//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint64 = 0x100000 // 1 MB
