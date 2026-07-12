// ASPEED AST2700 SoC support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ast2700 provides support to Go bare metal unikernels written using
// the TamaGo framework on the ASPEED AST2700 SoC (quad-core Cortex-A35).
//
// The package implements initialization and drivers for the AST2700 SoC,
// adopting the following reference specifications:
//   - AST2750 A2 Datasheet v1.11 (March 2026)
//
// Build with:
//
//	GOOS=tamago GOARCH=arm64
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package ast2700

import (
	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/arm64/gic"
	"github.com/usbarmory/tamago/soc/aspeed/uart"
)

// SiliconRevision holds the hardware revision read from SCUIO 0x000 bits[23:16].
// Populated during init(). Values: 0x00=A0, 0x01=A1, 0x02=A2.
var SiliconRevision uint32

// ARM processor instance.
var ARM = &arm64.CPU{}

// GIC is the ARM Generic Interrupt Controller (GICv3).
var GIC = &gic.GIC{
	GICD: GIC_DIST_BASE,
	GICR: GIC_REDIST_BASE,
}

// UART12 is the primary console UART (16550-compatible, 115200-8N1).
var UART12 = &uart.UART{
	Index: 12,
	Base:  UART12_BASE,
}
