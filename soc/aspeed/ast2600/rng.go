// ASPEED AST2600 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package ast2600

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

// ScuRng2CtrlDataValid is the bit position of the data-valid flag in RNG2_CTRL.
// When bit 31 of SCU530 is set, SCU534 holds a valid random word.
const ScuRng2CtrlDataValid = 31

// rngPollMax is the maximum number of polling iterations before falling back to
// software PRNG. On real AST2600 hardware the RNG2 produces data within a few
// cycles. On QEMU or uninitialized systems the bit may never be set.
const rngPollMax = 16

// rngLCG holds state for the Lehmer LCG software fallback.
// Seeded with a non-zero constant; real hardware seeds it from RNG2.
var rngLCG uint32 = 0xDEADBEEF

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	EnableHACEClock()
	// Enable RNG2 (bit0=0 means enabled).
	SCU.WriteRNG2CTRL(0)
	rng.GetRandomDataFn = getRandomData
}

func getRandomData(b []byte) {
	read := 0
	for read < len(b) {
		// Poll RNG2 data-valid flag with a bounded spin.
		// Falls back to a software LCG if the hardware doesn't respond
		// (e.g. under QEMU where the RNG peripheral is not emulated).
		var word uint32
		poll := 0
		for SCU.ReadRNG2CTRL()>>ScuRng2CtrlDataValid == 0 {
			poll++
			if poll >= rngPollMax {
				// Hardware RNG unavailable — advance the software LCG.
				rngLCG = rngLCG*1664525 + 1013904223
				word = rngLCG
				goto fill
			}
		}
		// Reading RNG2_DATA clears the data-valid flag.
		word = SCU.ReadRNG2DATA()
		// Seed the LCG with hardware entropy for future fallback calls.
		rngLCG ^= word
	fill:
		read = rng.Fill(b, read, word)
	}
}
