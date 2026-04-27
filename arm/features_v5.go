// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !arm.6

// Package arm provides support for ARM 32-bit architecture specific
// operations.
package arm

// initFeatures is a no-op on ARMv5 cores (GOARM=5): the CP15 ID_PFR0 and
// ID_PFR1 registers are not architecturally defined before ARMv6, so
// feature detection is skipped. The NoVBAR field and other CPU struct
// configuration must be set explicitly by the SoC package instead.
func (cpu *CPU) initFeatures() {}
