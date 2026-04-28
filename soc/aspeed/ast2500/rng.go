// ASPEED AST2500 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2500

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

// rngLCG holds state for the Lehmer LCG software PRNG.
// AST2500 has no hardware TRNG; a software fallback is used.
// Seed is mixed with timer counter on first call.
var rngLCG uint32 = 0xDEADBEEF
var rngSeeded bool

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	rng.GetRandomDataFn = getRandomData
}

func getRandomData(b []byte) {
	read := 0
	for read < len(b) {
		if !rngSeeded {
			// Seed LCG from timer counter for some entropy.
			rngLCG ^= TIMER.ReadT1STATUS()
			rngSeeded = true
		}
		rngLCG = rngLCG*1664525 + 1013904223
		read = rng.Fill(b, read, rngLCG)
	}
}
