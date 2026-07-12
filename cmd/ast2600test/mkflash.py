#!/usr/bin/env python3
# mkflash.py — create an ASPEED AST2600 FMC flash image
#
# Usage:
#   mkflash.py <input.bin> <output.img> [flash_size [entry_point]]
#
#   input.bin    raw binary produced by:
#                  arm-none-eabi-objcopy -O binary
#                    --remove-section=.note.go.pvh
#                    --remove-section=.note.go.buildid
#                    --remove-section=.note.gnu.build-id
#                  ast2600test.elf ast2600test.bin
#                Stripping the note sections ensures the binary starts at the
#                .text VMA (0x80010000) so it can be loaded there in DRAM.
#
#   entry_point  hex ELF entry address (e.g. 0x80095444); obtained from:
#                  arm-none-eabi-readelf -h ast2600test.elf | awk '/Entry/{print $4}'
#                Defaults to LOAD_ORIGIN (0x80010000) if not supplied.
#
# Flash layout:
#   0x000000 – 0x00006B  108-byte ARM loader stub (runs at 0x20000000)
#   0x00006C – 0x00FFFF  zero-padded (erased flash)
#   0x010000 – ...       TamaGo raw binary (binary starts here at flash level)
#   ...      – end       zero-padded to FLASH_SIZE (64 MB = w25q512jv default)
#
# Boot flow — HARDWARE:
#   1. AST2600 boot ROM initialises SDRAM then jumps to FMC CS0 (0x20000000).
#   2. CPU1 is parked in a WFE loop by the stub.  CPU0:
#      a. Compares the first word from 0x20010000 (flash) with
#         0x80010000 (DRAM).  On a cold boot DRAM is zeroed so the words
#         differ and the copy runs.
#      b. Bulk-copies the TamaGo binary from flash (0x20010000) to DRAM
#         (0x80010000) using 32-byte LDMIA/STMIA bursts.
#   3. Jumps to the ELF entry point (stored in the stub header at 0x0C).
#
# Boot flow — QEMU (ast2600-evb):
#   QEMU's FMC/SPI emulation is accurate but slow; a 1.5 MB bulk copy through
#   memory-mapped flash would take minutes.  The Makefile qemu-flash target
#   adds:
#       -device loader,force-raw=on,addr=0x80010000,file=ast2600test.bin
#       -device loader,cpu-num=0,addr=<entry>
#       -device loader,cpu-num=1,addr=<entry>
#   QEMU pre-loads the binary at 0x80010000 and both CPUs start at the entry
#   point directly, bypassing the flash boot path entirely (the flash stub
#   never runs in this mode).
#
# Stub assembly (arm-none-eabi-as -march=armv7-a verified, 108 bytes / 27 words):
#
#   0x20000000  EA000004  B loader             ; branch over header (→ 0x20000018)
#   0x20000004  474D4154  .word "TAMG"         ; magic
#   0x20000008  <size>    .word payload_size   ; patched: copy_size (32B-aligned)
#   0x2000000C  <entry>   .word entry_point    ; patched: ELF entry address
#   0x20000010  E320F002  cpu1_spin: WFE
#   0x20000014  EAFFFFFD  B cpu1_spin          ; (→ 0x20000010)
#   0x20000018  EE100FB0  loader: MRC p15,0,r0,c0,c0,5  ; read MPIDR
#   0x2000001C  E2000003  AND r0,r0,#3         ; isolate affinity bits[1:0]
#   0x20000020  E3500000  CMP r0,#0
#   0x20000024  1AFFFFF9  BNE cpu1_spin        ; non-zero → CPU1, park
#   0x20000028  E59F0028  LDR r0,[PC,#0x28]   ; r0 = src = 0x20010000
#   0x2000002C  E59F1028  LDR r1,[PC,#0x28]   ; r1 = dst = 0x80010000
#   0x20000030  E5903000  LDR r3,[r0]          ; first word from flash
#   0x20000034  E5914000  LDR r4,[r1]          ; first word from DRAM
#   0x20000038  E1530004  CMP r3,r4
#   0x2000003C  0A000004  BEQ skip_copy        ; equal → pre-loaded, skip
#   0x20000040  E59F2018  LDR r2,[PC,#0x18]   ; r2 = copy size
#   0x20000044  E8B007F8  LDMIA r0!,{r3-r10}  ; copy_loop: 32 bytes from flash
#   0x20000048  E8A107F8  STMIA r1!,{r3-r10}  ; store to DRAM
#   0x2000004C  E2522020  SUBS r2,r2,#32
#   0x20000050  CAFFFFFB  BGT copy_loop        ; → 0x20000044
#   0x20000054  E59FF008  skip_copy: LDR pc,[PC,#0x08]  ; jump → entry_point
#   0x20000058  20010000  src literal          ; flash + 64 KB
#   0x2000005C  80010000  dst literal          ; binary load base in DRAM
#   0x20000060  <size>    size literal         ; patched (same as 0x08)
#   0x20000064  <entry>   entry literal        ; patched (same as 0x0C)
#   0x20000068  00000000  padding

import struct
import sys

PAYLOAD_OFFSET = 0x10000          # binary at flash offset 64 KB
LOAD_ORIGIN    = 0x80010000       # DRAM address where binary is loaded
FLASH_SIZE     = 64 * 1024 * 1024  # 64 MB: matches QEMU ast2600-evb default (w25q512jv)

_STUB_TEMPLATE = [
    0xEA000004,  # [0x00] B loader                      (→ 0x20000018)
    0x474D4154,  # [0x04] "TAMG" magic
    0x00000000,  # [0x08] payload size  [index 2]   — patched
    0x00000000,  # [0x0C] entry point   [index 3]   — patched
    0xE320F002,  # [0x10] cpu1_spin: WFE
    0xEAFFFFFD,  # [0x14] B cpu1_spin                   (→ 0x20000010)
    0xEE100FB0,  # [0x18] loader: MRC p15,0,r0,c0,c0,5  (read MPIDR)
    0xE2000003,  # [0x1C] AND r0,r0,#3
    0xE3500000,  # [0x20] CMP r0,#0
    0x1AFFFFF9,  # [0x24] BNE cpu1_spin                 (→ 0x20000010)
    0xE59F0028,  # [0x28] LDR r0,[PC,#0x28]             (src = 0x20010000)
    0xE59F1028,  # [0x2C] LDR r1,[PC,#0x28]             (dst = 0x80010000)
    0xE5903000,  # [0x30] LDR r3,[r0]                   first word from flash
    0xE5914000,  # [0x34] LDR r4,[r1]                   first word from DRAM
    0xE1530004,  # [0x38] CMP r3,r4
    0x0A000004,  # [0x3C] BEQ skip_copy                 (→ 0x20000054)
    0xE59F2018,  # [0x40] LDR r2,[PC,#0x18]             (copy size)
    0xE8B007F8,  # [0x44] copy_loop: LDMIA r0!,{r3-r10}
    0xE8A107F8,  # [0x48] STMIA r1!,{r3-r10}
    0xE2522020,  # [0x4C] SUBS r2,r2,#32
    0xCAFFFFFB,  # [0x50] BGT copy_loop                 (→ 0x20000044)
    0xE59FF008,  # [0x54] skip_copy: LDR pc,[PC,#0x08]  jump → entry_point
    0x20010000,  # [0x58] src literal
    0x80010000,  # [0x5C] dst literal  (= LOAD_ORIGIN)
    0x00000000,  # [0x60] size literal  [index 24]  — patched
    0x00000000,  # [0x64] entry literal [index 25]  — patched
    0x00000000,  # [0x68] padding
]


def make_flash_image(payload_path, output_path,
                     flash_size=FLASH_SIZE, entry_point=LOAD_ORIGIN):
    with open(payload_path, "rb") as f:
        payload = f.read()

    # Copy loop works in 32-byte chunks; round up so the final iteration
    # does not stall.
    aligned_size = (len(payload) + 31) & ~31

    if PAYLOAD_OFFSET + aligned_size > flash_size:
        sys.exit(
            f"error: payload ({aligned_size:#x} B) does not fit at "
            f"offset {PAYLOAD_OFFSET:#x} in {flash_size:#x}-byte flash"
        )

    words = list(_STUB_TEMPLATE)
    words[2]  = aligned_size   # header: copy size
    words[3]  = entry_point    # header: ELF entry
    words[24] = aligned_size   # literal: copy size
    words[25] = entry_point    # literal: ELF entry

    stub = struct.pack(f"<{len(words)}I", *words)

    image = bytearray(flash_size)
    image[0:len(stub)] = stub
    image[PAYLOAD_OFFSET : PAYLOAD_OFFSET + len(payload)] = payload

    with open(output_path, "wb") as f:
        f.write(image)

    print(f"  image : {output_path}")
    print(f"  flash : {flash_size // (1024 * 1024)} MB")
    print(f"  binary: {len(payload):#x} B  ({len(payload) // 1024} KB)")
    print(f"  copy  : {aligned_size:#x} B  (32B-aligned; skipped if DRAM pre-loaded)")
    print(f"  src   : 0x20010000  (FMC CS0 + {PAYLOAD_OFFSET:#x})")
    print(f"  dst   : {LOAD_ORIGIN:#010x}  (binary load base in DRAM)")
    print(f"  entry : {entry_point:#010x}  (ELF entry point)")


if __name__ == "__main__":
    if len(sys.argv) not in (3, 4, 5):
        sys.exit(f"usage: {sys.argv[0]} <input.bin> <output.img> [flash_size [entry_point]]")
    fs    = int(sys.argv[3], 0) if len(sys.argv) >= 4 else FLASH_SIZE
    entry = int(sys.argv[4], 0) if len(sys.argv) == 5 else LOAD_ORIGIN
    make_flash_image(sys.argv[1], sys.argv[2], fs, entry)
