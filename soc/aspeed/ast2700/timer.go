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
)

// cpuFreqHz is the CPU counter frequency (CNTFRQ) in Hz.
// QEMU AST2700 sets this to 1125000000 (1.125 GHz).
const cpuFreqHz = 1125000000

// refFreq is the desired output frequency for nanotime (1 GHz = 1 ns per tick).
const refFreq int64 = 1e9

// timerMultiplier converts counter ticks to nanoseconds.
// = refFreq / cpuFreqHz = 1e9 / 1125000000 ≈ 0.8889
// We use integer math: nanoseconds = counter * refFreq / cpuFreqHz
const timerMultiplierNum = refFreq        // 1e9
const timerMultiplierDen = cpuFreqHz      // 1125000000

// defined in timer.s
func read_cntpct() uint64

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	// CNTPCT counts at cpuFreqHz. Convert to nanoseconds:
	//   ns = counter * 1e9 / cpuFreqHz
	//
	// A naive `counter * 1e9` overflows uint64 once the counter exceeds
	// 2^64/1e9 ≈ 1.845e10 ticks, i.e. after only ~16.4s at 1.125 GHz, which
	// made nanotime wrap to near-zero every ~16s and froze the Go runtime
	// scheduler. Split the counter into whole-period seconds plus a remainder
	// so neither term can overflow:
	//   sec = counter / cpuFreqHz                (uptime seconds)
	//   rem = counter % cpuFreqHz                (< cpuFreqHz)
	//   ns  = sec*1e9 + rem*1e9/cpuFreqHz
	// rem*1e9 < 1.125e9*1e9 = 1.125e18 < 2^64, so the remainder term is safe,
	// and sec*1e9 is the uptime in ns (fits int64 for ~292 years).
	c := read_cntpct()
	sec := c / uint64(timerMultiplierDen)
	rem := c % uint64(timerMultiplierDen)
	return int64(sec*uint64(timerMultiplierNum) + rem*uint64(timerMultiplierNum)/uint64(timerMultiplierDen))
}
