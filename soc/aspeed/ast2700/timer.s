// ASPEED AST2700 SoC support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#include "textflag.h"

// func read_cntpct() uint64
TEXT ·read_cntpct(SB),NOSPLIT,$0-8
	ISB	$0b1111
	MRS	CNTPCT_EL0, R0
	MOVD	R0, ret+0(FP)
	RET
