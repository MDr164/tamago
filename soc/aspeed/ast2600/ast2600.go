// ASPEED AST2600 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package ast2600 provides support to Go bare metal unikernels written using
// the TamaGo framework on the ASPEED AST2600 SoC (dual-core Cortex-A7).
//
// The package implements initialization and drivers for the AST2600 SoC,
// adopting the following reference specifications:
//   - AST2600 A3 Datasheet v1.6 (Jul 2024)
//
// This package is only meant to be used with
// `GOOS=tamago GOARCH=arm GOARM=7,softfloat -tags softfloat`
// as supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package ast2600

import (
	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/arm/gic"
)

// SiliconRevision holds the hardware revision read from SCU004 bits[23:16].
// Populated during init(). Values: 0x00=A0, 0x01=A1, 0x02=A2, 0x03=A3.
var SiliconRevision uint32

// ARM processor instance.
var ARM = &arm.CPU{}

// GIC is the ARM Generic Interrupt Controller (GICv2).
// Base is the private peripheral space base (GICD at Base+0x1000, GICC at Base+0x2000).
var GIC = &gic.GIC{
	Base: GIC_BASE,
}

// SCU is the System Control Unit peripheral accessor.
var SCU = &Scu{Base: uintptr(SCU_BASE)}

// UART5 is the primary console UART (16550-compatible).
var UART5 = &Uart{Base: uintptr(UART5_BASE)}

// TIMER is the Timer Controller peripheral accessor.
var TIMER = &Timer{Base: uintptr(TIMER_BASE)}
