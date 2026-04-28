// ARM processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !arm.6 || softfloat

package arm

import (
	_ "unsafe"
)

// Init takes care of the lower level initialization triggered before runtime
// setup (pre World start).
//
// On GOARM=5 (soft-float ABI) or softfloat builds (GOARM=6,softfloat /
// GOARM=7,softfloat with -tags softfloat) no VFP initialization is required
// at pre-World start. Cores without VFP (e.g. ARM1176JZS on AST2500) would
// fault on vfp_enable; softfloat Go code never issues VFP instructions.
//
//go:nosplit
//go:linkname Init runtime/goos.Hwinit0
func Init() {}
