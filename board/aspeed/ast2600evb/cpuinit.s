// ASPEED AST2600EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Board-specific CPU initialization for the AST2600EVB (dual-core Cortex-A7).
//
// The ASPEED boot ROM initializes SDRAM and loads the TamaGo FMC image to
// 0x80000000 before jumping here. Both Cortex-A7 CPUs start at the entry
// point. Only CPU0 proceeds with initialization; CPU1 spins in a low-power
// WFE loop until explicitly released via goos.Task.
//
// Entry state (from ASPEED boot ROM): SVC mode, IRQ/FIQ disabled, SDRAM ready.

//go:build linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Read MPIDR (Multiprocessor Affinity Register) into R0.
	// MRC p15, 0, r0, c0, c0, 5 = 0xEE100FB0
	// Bits [1:0] = CPU affinity level 0 (0=CPU0, 1=CPU1).
	WORD	$0xEE100FB0
	AND	$0x3, R0, R0
	CMP	$0, R0
	B.NE	cpu_spin		// CPU1 spins here

	// Compute initial stack pointer: RamStart + RamSize - RamStackOffset
	MOVW	runtime∕goos·RamStart(SB), R13
	MOVW	runtime∕goos·RamSize(SB), R1
	MOVW	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, R13
	SUB	R2, R13
	MOVW	R13, R3		// save for after mode switch

	// Detect current processor mode.
	WORD	$0xe10f0000	// mrs r0, CPSR
	AND	$0x1f, R0, R0

	// If already in USR mode just start the runtime (unusual).
	CMP	$0x10, R0
	BL.EQ	_rt0_tamago_start(SB)

	// Handle HYP mode (entered when loaded via certain boot paths).
	// Use ERET to transition to SVC mode with interrupts masked.
	CMP	$0x1a, R0
	B.NE	after_eret

	BIC	$0x1f, R0
	ORR	$0x1d3, R0		// SVC mode, AIF masked
	MOVW	$12(R15), R14		// lr = PC + 12 = after_eret
	WORD	$0xe16ff000		// msr SPSR_fsxc, r0
	WORD	$0xe12ef30e		// msr ELR_hyp, lr
	WORD	$0xe160006e		// eret

after_eret:
	// Switch to System mode (0xdf) to enable full register access.
	WORD	$0xe321f0df		// msr CPSR_c, #0xdf

	MOVW	R3, R13			// restore stack pointer

	// Zero BSS and no-pointer-BSS sections.
	BL	clearBSS(SB)

	B	_rt0_tamago_start(SB)

// cpu_spin: secondary CPUs wait in a low-power WFE loop.
// They remain here until TamaGo's SMP support (via goos.Task) releases them.
// SEV from the primary CPU will wake them when SMP is enabled.
cpu_spin:
	WORD	$0xe320f002		// wfe (ARMv7 WFE opcode)
	B	cpu_spin

// clearBSS zeroes the .bss and .noptrbss sections.
TEXT clearBSS(SB),NOSPLIT|NOFRAME,$0
	MOVW	$runtime·bss(SB), R0
	MOVW	$runtime·enoptrbss(SB), R1
	MOVW	$0, R2
bss_loop:
	CMP	R0, R1
	B.EQ	bss_done
	MOVW	R2, (R0)
	ADD	$4, R0
	B	bss_loop
bss_done:
	RET
