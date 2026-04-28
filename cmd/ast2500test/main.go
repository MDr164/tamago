// ASPEED AST2500 TamaGo boot test
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
	"time"

	_ "github.com/usbarmory/tamago/board/aspeed/ast2500evb"
	"github.com/usbarmory/tamago/soc/aspeed/ast2500"
)

func main() {
	fmt.Printf("Hello from TamaGo on ASPEED AST2500!\n\n")
	fmt.Printf("Runtime : %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Board   : AST2500EVB\n")
	fmt.Printf("SoC     : %s\n", ast2500.Model())
	fmt.Printf("CPU     : ARM1176JZS\n")
	fmt.Printf("RAM     : 512 MB @ 0x80000000 (remapped to 0x0)\n")

	for i := 0; ; i++ {
		fmt.Printf("tick %d\n", i)
		time.Sleep(1 * time.Second)
	}
}
