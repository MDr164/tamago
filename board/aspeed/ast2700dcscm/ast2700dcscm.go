// ASPEED AST2700 DCSCM board support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ast2700dcscm provides hardware initialization, automatically on
// import, for the AST2750-A1 DCSCM DDR4 board.
//
// The BootMCU initializes DRAM, loads the TamaGo payload, and releases the
// CA35 core. This package handles only CA35-side board-level setup.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go on ARM64 SoCs, see
// https://github.com/usbarmory/tamago.
package ast2700dcscm

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/aspeed/ast2700"
)

// Peripheral instances re-exported for application convenience.
var (
	UART12 = ast2700.UART12
	GIC    = ast2700.GIC
	ARM    = ast2700.ARM
)

// Init performs board-level hardware initialization triggered early in
// runtime setup (post Go-World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	// DEBUG breadcrumb 'E': reached Hwinit1 — this runs AFTER the runtime's
	// Hwinit0 (arm64 InitMMU, which enables the MMU + I/D caches) and after the
	// Go world has started (so atomics/scheduler work). If 'E' prints, the MMU
	// and cache enable survived; if it does not, execution died in InitMMU
	// (uninvalidated reset cache lines on real silicon).
	ast2700.UART12.Tx('E')
	ast2700.Init()
	ast2700.UART12.Init()
	ast2700.UART12.Tx('F')
}
