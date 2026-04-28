// ASPEED AST2500EVB board support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Board-specific CPU initialization for the AST2500EVB (ARM1176JZS, ARMv6).
//
// Linker text base: 0x80010000 (canonical SDRAM so QEMU loads the ELF correctly).
// goos.RamStart = 0x0: after SDRAM remap the exception vectors and Go runtime
// structures live at 0x0-based addresses. Both 0x0 and 0x80000000 map to the
// same physical SDRAM after EarlyInit sets AHBC8C[0]=1.
//
// ARM1176JZS specifics:
//   - NoVBAR=true: no VBAR register; exception stubs must be at 0x00000000
//   - ARMv6 (arm.6 build tag): CPSIE/CPSID and WFI instruction available
//   - GOARM=6,softfloat -tags softfloat: ARM1176JZS has no VFP unit

//go:build linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Step 1: SoC early init — sets AHBC8C[0]=1 (SDRAM→0x0 remap),
	// unlocks SCU, enables UART5 clock. No stack used.
	BL	github·com∕usbarmory∕tamago∕soc∕aspeed∕ast2500·EarlyInit(SB)

	// Step 2: Install ARM exception vector stubs at 0x00000000.
	// After AHBC8C remap, 0x0 aliases SDRAM and is writable.
	// Each slot = 0xEAFFFFFE (B . — branch to self). ARM.Init() will
	// overwrite these with proper LDR PC handlers during Hwinit1.
	MOVW	$0xEAFFFFFE, R0
	MOVW	$0, R1
	MOVW	$8, R2
stub_loop:
	MOVW	R0, (R1)
	ADD	$4, R1
	SUB	$1, R2
	CMP	$0, R2
	B.NE	stub_loop

	// Step 3: Compute stack pointer from goos.RamStart (= 0x0 post-remap).
	MOVW	runtime∕goos·RamStart(SB), R13
	MOVW	runtime∕goos·RamSize(SB), R1
	MOVW	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, R13
	SUB	R2, R13
	MOVW	R13, R3		// save for after mode switch

	// Step 4: Detect current mode.
	WORD	$0xe10f0000	// mrs r0, CPSR
	AND	$0x1f, R0, R0

	// USR mode → jump directly (rare path).
	CMP	$0x10, R0
	BL.EQ	_rt0_tamago_start(SB)

	// No HYP mode on ARM1176; skip HYP handling.
	// Switch to System mode (0xdf = SYS + I+F masked).
	WORD	$0xe321f0df		// msr CPSR_c, #0xdf

	MOVW	R3, R13			// restore stack pointer

	// Step 5: Zero BSS.
	BL	clearBSS(SB)

	B	_rt0_tamago_start(SB)

// clearBSS zeroes .bss and .noptrbss sections.
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
