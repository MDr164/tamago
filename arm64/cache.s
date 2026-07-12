// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "arm64.h"

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D12.2.100 SCTLR_EL1, System Control Register (EL1)

// func cache_disable()
TEXT ·cache_disable(SB),$0
	MRS	SCTLR_EL1, R0
	BIC	$1<<12, R0	// disable I-cache
	BIC	$1<<2, R0	// disable D-cache
	MSR	R0, SCTLR_EL1
	ISB	SY
	RET

// func cache_enable()
TEXT ·cache_enable(SB),$0
	MRS	SCTLR_EL1, R0
	ORR	$1<<12, R0	// enable I-cache
	ORR	$1<<2, R0	// enable D-cache
	MSR	R0, SCTLR_EL1
	ISB	SY
	RET

// func cache_clean_range(start, size uintptr)
//
// DC CVAC: Clean data cache by VA to Point of Coherency.
// Writes dirty cache lines back to memory without invalidating them.
TEXT ·cache_clean_range(SB),$0-16
	MOVD	start+0(FP), R0
	MOVD	size+8(FP), R1
	ADD	R0, R1, R1	// R1 = end
	// Align start down to cache line boundary (64 bytes)
	AND	$~63, R0
clean_loop:
	CMP	R1, R0
	BGE	clean_done
	// DC CVAC, X0 — encoded as SYS instruction
	WORD	$0xd50b7a20
	ADD	$64, R0
	B	clean_loop
clean_done:
	DSB	ISH
	RET

// func cache_invalidate_range(start, size uintptr)
//
// DC CIVAC: Clean and Invalidate data cache by VA to Point of Coherency.
// Flushes dirty lines and discards cached copies so next read comes from memory.
TEXT ·cache_invalidate_range(SB),$0-16
	MOVD	start+0(FP), R0
	MOVD	size+8(FP), R1
	ADD	R0, R1, R1	// R1 = end
	// Align start down to cache line boundary (64 bytes)
	AND	$~63, R0
inv_loop:
	CMP	R1, R0
	BGE	inv_done
	// DC CIVAC, X0 — encoded as SYS instruction
	WORD	$0xd50b7e20
	ADD	$64, R0
	B	inv_loop
inv_done:
	DSB	ISH
	RET
