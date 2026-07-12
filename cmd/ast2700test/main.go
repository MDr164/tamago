// ASPEED AST2700 TamaGo boot test
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"runtime"
	goos "runtime/goos"

	"github.com/usbarmory/tamago/soc/aspeed/ast2700"
)

func main() {
	fmt.Printf("Hello World\n")
	fmt.Printf("Runtime      : %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("SoC          : %s %s\n", ast2700.Model(), ast2700.Revision())
	fmt.Printf("Raw revision : 0x%08x\n", ast2700.RawRevision())
	fmt.Printf("Generation   : 0x%02x\n", ast2700.Generation())
	fmt.Printf("Device ID    : 0x%02x\n", ast2700.DeviceID())
	fmt.Printf("HW revision  : 0x%02x\n", ast2700.HardwareRevision())
	fmt.Printf("CPU          : Cortex-A35\n")
	fmt.Printf("RAM          : %d MB @ 0x%x\n", goos.RamSize>>20, goos.RamStart)

	select {}
}
