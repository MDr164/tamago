// Raspberry Pi 2 Support
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package pi2

import (
	// Using go:linkname
	_ "unsafe"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x00000000

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x40000000 - 0x4C00000 // 1GB - 76MB (VideoCore)
