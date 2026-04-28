// ASPEED AST2600EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package ast2600evb

import _ "unsafe"

// The AST2600EVB evaluation board has 512 MB DDR4 SDRAM at 0x80000000.
// Applications that need a larger or smaller heap can override this with
// -tags linkramsize and their own mem.go.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint32 = 0x20000000 // 512 MB
