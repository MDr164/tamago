// ASPEED AST2500 SoC support for tamago/arm
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// EarlyInit performs minimal SoC initialization that must happen before the Go
// runtime starts. Runs in pure assembly with no stack, no heap, no Go runtime.
//
// Actions:
//   1. Set AHBC8C[0]=1 (0x1E60008C) to remap SDRAM to 0x00000000
//      (canonical SDRAM at 0x80000000 remains accessible at both addresses)
//   2. Unlock SCU (write 0x1688A8A8 to SCU00 = 0x1E6E2000)
//   3. Enable UART5 clock (clear bit 15 in SCU0C = 0x1E6E200C)
//   4. Lock SCU
//
// Called from board cpuinit.s via BL ·EarlyInit(SB).

#include "textflag.h"

// func EarlyInit()
TEXT ·EarlyInit(SB),NOSPLIT|NOFRAME,$0
	// Remap SDRAM to 0x00000000 via AHBC8C (AHB Controller, offset 0x8C).
	// AHBC base = 0x1E600000. AHBC8C bit 0 = boot area remap:
	//   0 = 0x0000_0000 maps to static memory (SPI flash)
	//   1 = 0x0000_0000 maps to SDRAM
	// This allows exception vectors and Go heap to live at 0x0-based addresses
	// while canonical SDRAM at 0x80000000 remains accessible for code.
	MOVW	$0x1E60008C, R0
	MOVW	(R0), R1
	ORR	$1, R1
	MOVW	R1, (R0)

	// Unlock SCU: write 0x1688A8A8 to SCU base (0x1E6E2000)
	MOVW	$0x1E6E2000, R0
	MOVW	$0x1688A8A8, R1
	MOVW	R1, (R0)

	// Enable UART5 clock: clear bit 15 of SCU0C (offset 0x0C from base)
	MOVW	$0x1E6E200C, R0
	MOVW	(R0), R1
	BIC	$(1<<15), R1
	MOVW	R1, (R0)

	// Enable Timer clock: Timer uses PCLK; no dedicated clock gate in AST2500
	// (Timer is enabled by its control register, no SCU clock gate needed)

	// Lock SCU: write non-unlock value to SCU00
	MOVW	$0x1E6E2000, R0
	MOVW	$0x1, R1
	MOVW	R1, (R0)

	RET
