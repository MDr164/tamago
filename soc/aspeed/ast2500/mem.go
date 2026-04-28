// ASPEED AST2500 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramstart

package ast2500

import _ "unsafe"

// SDRAM base for AST2500. Canonical physical address is 0x80000000.
// The ASPEED boot ROM may remap SDRAM to 0x00000000 via AHBC8C[0], but
// TamaGo uses the canonical 0x80000000 base so that:
//   - QEMU loading with -kernel places the ELF in real SDRAM
//   - The stack lands in valid SDRAM (0x9FF00000 for 512 MB)
// On real hardware the EarlyInit sets the AHBC remap so that exception
// vectors installed at 0x0 (by cpuinit.s) in remapped SDRAM are accessible.
// The arm.CPU struct has NoVBAR=true so VBAR is not written; the LDR PC table
// written by initVectorTable goes to RamStart (0x80000000 in SDRAM).
//
//go:linkname ramStart runtime/goos.RamStart
var ramStart uint32 = 0x80000000

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB
