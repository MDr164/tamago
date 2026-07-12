// QEMU AST2700 support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ast2700 provides hardware initialization, automatically on import,
// for the QEMU ast2700a2-evb machine configured with a single Cortex-A35 core.
//
// Build with:
//
//	GOOS=tamago GOARCH=arm64 \
//	  go tool tamago build \
//	    -ldflags "-T 0x400010000 -R 0x1000" \
//	    ./cmd/myapp
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go on ARM64 SoCs, see
// https://github.com/usbarmory/tamago.
package ast2700

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
	ast2700.Init()
	ast2700.UART12.Init()
}
