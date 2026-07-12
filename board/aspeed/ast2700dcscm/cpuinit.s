// ASPEED AST2700 DCSCM board support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Board-specific CPU initialization for the AST2750-A1 DCSCM DDR4 board.
//
// The BootMCU releases the CA35 at EL3. This code:
//   1. Parks secondary CPUs (WFE loop).
//   2. Configures EL3 system registers (FP/SIMD, timers, GICv3 SRE).
//   3. Configures EL2 to pass timer/FP access down to EL1.
//   4. Drops CPU0 to Secure EL1h via ERET.
//   5. Sets the stack from runtime symbols and enters _rt0_tamago_start.

//go:build linkcpuinit

#include "textflag.h"

TEXT cpuinit(SB),NOSPLIT|NOFRAME,$0
	// Park secondary CPUs.
	MRS	MPIDR_EL1, R0
	AND	$0xff, R0, R0
	CMP	$0, R0
	BNE	cpu_spin

	// DEBUG: liveness handshake via DRAM scratch — CA35 view 0x4_03ff_e000
	// == BootMCU view 0x83ffe000. Written before any UART/IO-die access so
	// the BootMCU can tell "CA35 executed" from "CA35 can't reach UART".
	MOVD	$0x403ffe000, R9
	MOVD	$0xCA35A11E, R10
	MOVW	R10, (R9)

	// DEBUG breadcrumb 'A': CA35 reached cpuinit (reset EL, before any msr).
	MOVD	$0x14c33b00, R9
	MOVD	$0x41, R10
	MOVB	R10, (R9)

	// Enable SMP coherency: CPUECTLR_EL1.SMPEN (S3_1_C15_C2_1, bit 6). This
	// MUST be set before the caches/MMU are enabled (done later in the arm64
	// runtime InitMMU). On the Cortex-A35 the core does not participate in the
	// Inner Shareable coherency domain until SMPEN is set, so exclusive
	// (LDXR/STXR) monitors never succeed and the Go runtime's atomic self-test
	// (runtime.check -> testAtomic64) livelocks. QEMU does not model this, which
	// is why the hang only appears on real silicon. Matches ATF cortex_a35.S.
	WORD	$0xD539F220	// mrs x0, S3_1_c15_c2_1 (CPUECTLR_EL1)
	ORR	$1<<6, R0	// set SMPEN
	WORD	$0xD519F220	// msr S3_1_c15_c2_1, x0
	ISB	$0b1111

	// DEBUG: read CPUECTLR_EL1 back and report SMPEN (bit 6) as 'Y'/'N' so we
	// can confirm the write actually stuck at this exception level.
	WORD	$0xD539F220	// mrs x0, S3_1_c15_c2_1 (CPUECTLR_EL1)
	AND	$1<<6, R0, R11
	MOVD	$0x14c33b00, R9
	CMP	$0, R11
	BEQ	smpen_off
	MOVD	$0x59, R10	// 'Y'
	MOVB	R10, (R9)
	B	smpen_done
smpen_off:
	MOVD	$0x4E, R10	// 'N'
	MOVB	R10, (R9)
smpen_done:

	// ── EL3 system register setup ─────────────────────────────────────

	// CPTR_EL3: clear TFP (bit 10) to allow FP/SIMD at all ELs.
	MOVD	$0, R0
	WORD	$0xd51e1140	// msr CPTR_EL3, x0

	// CNTFRQ_EL0: counter frequency (1.125 GHz on AST2700).
	MOVD	$1125000000, R0
	MSR	R0, CNTFRQ_EL0

	// Enable the system counter (CNTCR.EN at 0x12C10000).
	MOVD	$0x12c10000, R3
	MOVD	$1, R4
	MOVW	R4, (R3)

	// SCR_EL3: RW=1 (AArch64 at EL2/EL1), HCE=1, RES1 bits 4,5.
	// NS=0: stay secure so EL1 can access IO-die MMIO without TZASC.
	MOVD	$0, R0
	ORR	$1<<10, R0	// RW
	ORR	$1<<8, R0	// HCE
	ORR	$1<<5, R0	// RES1
	ORR	$1<<4, R0	// RES1
	WORD	$0xd51e1100	// msr SCR_EL3, x0

	// ── EL2 passthrough setup ─────────────────────────────────────────

	// HCR_EL2: RW=1 (EL1 is AArch64).
	MOVD	$1<<31, R0
	WORD	$0xd51c1100	// msr HCR_EL2, x0

	// CNTHCTL_EL2: allow EL1/EL0 physical timer and counter access.
	MOVD	$3, R0		// EL1PCEN | EL1PCTEN
	WORD	$0xd51ce100	// msr CNTHCTL_EL2, x0

	// CNTVOFF_EL2: zero virtual offset.
	MOVD	$0, R0
	WORD	$0xd51ce060	// msr CNTVOFF_EL2, x0

	// CPTR_EL2: allow FP/SIMD at EL2.
	MOVD	$0, R0
	WORD	$0xd51c1140	// msr CPTR_EL2, x0

	// ICC_SRE_EL3: enable GICv3 system register interface for all ELs.
	// EN (bit 3) | DIB (bit 2) | DFB (bit 1) | SRE (bit 0)
	MOVD	$0xf, R0
	WORD	$0xd51ecca0	// msr ICC_SRE_EL3, x0
	ISB	$0b1111

	// ── Drop to Secure EL1h ──────────────────────────────────────────

	// SPSR_EL3: all exceptions masked, target EL1h.
	MOVD	$0, R0
	ORR	$0b1111<<6, R0	// DAIF masked
	ORR	$0b0101<<0, R0	// EL1h (M[3:0]=0101)
	WORD	$0xd51e4000	// msr SPSR_EL3, x0

	MOVD	$·cpuinit_el1(SB), R0
	WORD	$0xd51e4020	// msr ELR_EL3, x0
	ISB	$0b1111

	// DEBUG breadcrumb 'B': EL3 setup completed, about to ERET to EL1.
	MOVD	$0x14c33b00, R9
	MOVD	$0x42, R10
	MOVB	R10, (R9)

	ERET

cpu_spin:
	WFE
	B	cpu_spin

// cpuinit_el1 runs at Secure EL1 after the ERET from EL3.
TEXT ·cpuinit_el1(SB),NOSPLIT|NOFRAME,$0
	// DEBUG breadcrumb 'C': reached Secure EL1 after ERET.
	MOVD	$0x14c33b00, R9
	MOVD	$0x43, R10
	MOVB	R10, (R9)

	// DEBUG: install a minimal EL1 synchronous-exception vector at a 2KB-aligned
	// DRAM scratch (VBASE = 0x4_0000_2000, free between the runtime vector table
	// at RamStart and the MMU page tables at RamStart+0x4000). The runtime does
	// not program VBAR_EL1 until Hwinit1 (after runtime.check), so without this a
	// synchronous fault in check's first atomic vectors to an unset VBAR and
	// hangs silently. Entry at VBASE+0x200 (current EL, SP_ELx, Synchronous)
	// loads the absolute handler address and branches to it.
	MOVD	$0x400002200, R2
	MOVD	$0x58000052, R3		// ldr x18, #8
	MOVW	R3, (R2)
	MOVD	$0xd61f0240, R3		// br x18
	MOVW	R3, 4(R2)
	MOVD	$·dbg_exc(SB), R3	// absolute handler address
	MOVD	R3, 8(R2)
	MOVD	$0x400002000, R1	// VBASE (2KB aligned)
	WORD	$0xd518c001		// msr vbar_el1, x1
	ISB	$0b1111

	// CPACR_EL1: FPEN=0b11 — allow FP/SIMD at EL1/EL0.
	MRS	CPACR_EL1, R0
	ORR	$3<<20, R0
	MSR	R0, CPACR_EL1
	ISB	$0b1111

	// Disable alignment check and MMU (clean state for hwinit0/InitMMU).
	MRS	SCTLR_EL1, R0
	BIC	$1<<1, R0	// A
	BIC	$1<<0, R0	// M
	MSR	R0, SCTLR_EL1
	ISB	$0b1111

	// Set stack pointer from runtime symbols.
	MOVD	runtime∕goos·RamStart(SB), R1
	MOVD	R1, RSP
	MOVD	runtime∕goos·RamSize(SB), R1
	MOVD	runtime∕goos·RamStackOffset(SB), R2
	ADD	R1, RSP
	SUB	R2, RSP

	// DEBUG breadcrumb 'D': stack set from RamStart/RamSize, entering _rt0.
	MOVD	$0x14c33b00, R9
	MOVD	$0x44, R10
	MOVB	R10, (R9)

	B	_rt0_tamago_start(SB)

// dbg_exc is a minimal, self-contained EL1 synchronous-exception handler used
// during CA35 bring-up. It writes 'X' followed by the two-hex-digit ESR_EL1.EC
// (exception class) to UART12, then spins — turning an otherwise-silent early
// fault into visible output that identifies the cause. No stack, FP, atomics or
// runtime state are used, so it is safe before the Go world is up.
TEXT ·dbg_exc(SB),NOSPLIT|NOFRAME,$0
	MOVD	$0x14c33b00, R9
	MOVD	$0x58, R10	// 'X'
	MOVB	R10, (R9)

	WORD	$0xd5385200	// mrs x0, esr_el1
	LSR	$26, R0, R0
	AND	$0x3f, R0, R0

	// high nibble of EC
	LSR	$4, R0, R11
	AND	$0xf, R11, R11
	ADD	$0x30, R11, R11
	CMP	$0x3a, R11
	BLT	dbg_hi_ok
	ADD	$0x27, R11, R11
dbg_hi_ok:
	MOVB	R11, (R9)

	// low nibble of EC
	AND	$0xf, R0, R11
	ADD	$0x30, R11, R11
	CMP	$0x3a, R11
	BLT	dbg_lo_ok
	ADD	$0x27, R11, R11
dbg_lo_ok:
	MOVB	R11, (R9)

	MOVD	$0x0D, R10
	MOVB	R10, (R9)
	MOVD	$0x0A, R10
	MOVB	R10, (R9)

dbg_exc_spin:
	B	dbg_exc_spin
