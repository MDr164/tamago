// ASPEED AST2700 SoC support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2700

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

// rngLCG holds state for the Lehmer LCG software fallback.
var rngLCG uint32 = 0xDEADBEEF

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	rng.GetRandomDataFn = getRandomData
}

func getRandomData(b []byte) {
	read := 0
	for read < len(b) {
		// Use software LCG fallback. On real hardware the TRNG at
		// TRNG_BASE would provide entropy, but QEMU does not emulate it.
		rngLCG = rngLCG*1664525 + 1013904223
		read = rng.Fill(b, read, rngLCG)
	}
}
