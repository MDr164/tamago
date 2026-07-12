// QEMU AST2700 support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !linkramsize

package ast2700

import _ "unsafe"

// The QEMU ast2700a2-evb machine defaults to 1 GB DRAM.
// Applications that need a larger or smaller heap can override this with
// -tags linkramsize and their own mem.go.

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x80000000 // 2 GB
