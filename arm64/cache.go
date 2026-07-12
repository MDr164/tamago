// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

// defined in cache.s
func cache_enable()
func cache_disable()
func cache_clean_range(start, size uintptr)
func cache_invalidate_range(start, size uintptr)

// EnableCache activates the ARM instruction and data caches.
func (cpu *CPU) EnableCache() {
	cache_enable()
}

// DisableCache disables the ARM instruction and data caches.
func (cpu *CPU) DisableCache() {
	cache_disable()
}

// FlushTLBs flushes the ARM Translation Lookaside Buffers.
func (cpu *CPU) FlushTLBs() {
	flush_tlb()
}

// CleanDataCacheRange cleans (writes back) data cache lines covering
// [start, start+size) to the Point of Coherency. Use before DMA reads
// from memory written by the CPU.
func (cpu *CPU) CleanDataCacheRange(start, size uintptr) {
	cache_clean_range(start, size)
}

// InvalidateDataCacheRange cleans and invalidates data cache lines
// covering [start, start+size). Use before the CPU reads memory
// written by a DMA controller.
func (cpu *CPU) InvalidateDataCacheRange(start, size uintptr) {
	cache_invalidate_range(start, size)
}
