# ROADMAP: ASPEED AST-Line TamaGo Support

> **Branch:** `aspeed` (contains merged NUC980/ARMv5 foundation from PR #66)
> **Spec:** `SPEC.md` — consult for boundaries, style conventions, and success criteria
> **Datasheets:** `/home/mdr164/private/kyanite/docs/md/ast{2400,2500,2600,2700}*.md`
> **U-Boot reference:** `/home/mdr164/private/kyanite/u-boot/` (branch `aspeed-master-v2023.10`)
> **chiptool:** `/home/mdr164/private/kyanite/chiptool/`

## Quick-reference: Confirmed Hardware Addresses

| SoC | CPU | `GOARCH`/`GOARM` | DRAM base | Linker `-T` | Interrupt ctrl | Timer base | UART5 base | SCU base |
|---|---|---|---|---|---|---|---|---|
| AST2400 | ARM926EJ-S (ARMv5) | `arm`/`5` | `0x40000000` | `0x40010000` | VIC `0x1e6c0000` | `0x1e782000` | `0x1e784000` | `0x1e6e2000` |
| AST2500 | ARM1176JZS (ARMv6) | `arm`/`6,softfloat` | `0x80000000` | `0x80010000` | VIC `0x1e6c0000` | `0x1e782000` | `0x1e784000` | `0x1e6e2000` |
| AST2600 | Cortex-A7 (ARMv7) | `arm`/`7,softfloat` | `0x80000000` | `0x80010000` | GIC `0x40460000`¹ | ARM generic | `0x1e784000` | `0x1e6e2000` |
| AST2700 | Cortex-A35 (AArch64) | `arm64`/— | `0x400000000` | `0x400010000` | GICv3 | `0x12c10000` | `0x12c1a000`² | SCU0 `0x12c02000` |

¹ GICD at `0x40461000` (DTS); `arm/gic` adds `GICD_OFF=0x1000`, so `Base=0x40460000`. Confirm from AST2600 datasheet §Memory Map.
² AST2700 UART4 on SoC0. SoC1 UARTs live at `0x14c33000+`. Console defaults to UART4 SoC0; confirm from AST2700 datasheet.

**SCU protection key** (AST2400/2500/2600): write `0x1688A8A8` to SCU base offset `+0x000`.
**Silicon revision registers:** AST2400/2500 → SCU`+0x07C`; AST2600 → SCU`+0x014`; AST2700 → SCU0`+0x000`.
**CNTFRQ (AST2600):** read `SCU_HW_STRAP1` (`0x1e6e2500`) bits [10:8], map to CPU clock: `0b000/0b010`→1.2 GHz, `0b001/0b011`→1.6 GHz, `0b100+`→800 MHz; write value to `CNTFRQ` via `MCR p15,0,r0,c14,c0,0`.

---

## Dependency Graph

```
T0.1 (chiptool CLI) ──────────────────────────────┐
T0.2 (arm ARM11 fix) ──┐                           │
                        │                           ▼
                        ├──► T1.1 (AST2600 YAML) ──► T1.2 (ast2600 SoC) ──► T1.3 (board) ──► T1.4 (QEMU) ──► T1.5 (HW)
                        │
                        ├──► T2.1 (shared uart) ─┐
                        │    T2.2 (shared timer) ─┤──► T2.4 (AST2500 YAML) ──► T2.5 (ast2500 SoC) ──► T2.6 (board) ──► T2.7 (QEMU+HW)
                        │    T2.3 (shared intc)  ─┘          │
                        │                                     └──► T3.1 (AST2400 YAML) ──► T3.2 (ast2400 SoC) ──► T3.3 (board) ──► T3.4 (QEMU+HW)
                        │
                        └──► T4.1 (AST2700 addr) ──► T4.2 (AST2700 YAML) ──► T4.3 (ast2700 SoC) ──► T4.4 (board) ──► T4.5 (HW)
```

`[P]` marks tasks that can run in parallel with no blocking dependency.

---

## Phase 0: Tooling Foundations

> These two tasks are independent of each other (`[P]`) and block everything downstream.
> Neither requires hardware.

---

### Task 0.1 [P]: chiptool Go generator — CLI and output-path optimizations

**Description:** Add `--package` and `--reg-import` CLI flags to the chiptool Go backend and change output file naming so generated files land flat in the target SoC package directory with a `.gen.go` suffix, not in a `pac/` subdirectory.

**Why now:** Every chiptool YAML task in Phases 1–4 depends on this being correct before the first `generate` run. Changing the output format after files exist forces mass-rename.

**Acceptance criteria:**
- [ ] Running `chiptool generate --language go --package ast2600 --reg-import github.com/usbarmory/tamago/internal/reg --output /tmp/out input.yaml` produces `/tmp/out/uart.gen.go` (not `/tmp/out/pac/uart.go`) with `package ast2600` at the top.
- [ ] The device output file is `/tmp/out/device.gen.go` (not `device_<name>.go`).
- [ ] All generated files import `"github.com/usbarmory/tamago/internal/reg"` when `--reg-import` is provided.
- [ ] Default behavior (no `--package`, no `--reg-import`) is unchanged for existing Rust callers.
- [ ] `cargo test` in the chiptool repo passes.

**Verification:**
```bash
cd /home/mdr164/private/kyanite/chiptool
cargo build
./target/debug/chiptool generate --language go \
  --package ast2600 \
  --reg-import github.com/usbarmory/tamago/internal/reg \
  --output /tmp/gentest \
  tests/data/example.yaml   # or any valid YAML
ls /tmp/gentest/             # must NOT contain pac/
head -2 /tmp/gentest/*.gen.go  # must show 'package ast2600'
```

**Dependencies:** None

**Files to change (chiptool repo):**
- `src/generate/go/mod.rs` — change line ~55: `format!("pac/{}.go", ...)` → `format!("{}.gen.go", ...)`; change line ~62: `format!("device_{}.go", ...)` → `"device.gen.go".to_string()`
- `src/commands/generate.rs` (or wherever CLI args live) — add `--package <name>` and `--reg-import <path>` args; construct `GoOptions` from them instead of `GoOptions::default()`

**Size:** S

---

### Task 0.2 [P]: arm package — ARM11 (ARMv6) compatibility and `Init()` signature

**Description:** Fix two issues introduced by the NUC980 branch that would silently mis-behave for ARM1176 (ARMv6): the `initFeatures()` guard being coupled to `NoVBAR` at runtime, and the `Init(vbar uint32)` signature breaking existing callers.

**Background:**
- `arm/arm.go` (aspeed branch) has `if !cpu.NoVBAR { cpu.initFeatures() }`. For AST2500 (ARM1176, `NoVBAR=true`), this skips ID_PFR0/ID_PFR1 detection even though ARM1176 has those registers. The `security` field (TrustZone) would stay `false` and `cpu.Secure()` returns incorrect results.
- `arm/arm.go` changed `Init()` to `Init(vbar uint32)`. The NUC980 calls `ARM.Init(ramStart)` but all existing SoC packages (`imx6ul`, `imx8mp`, `bcm2835`, etc.) call `ARM.Init()` with no argument — compile error on those targets.

**Acceptance criteria:**
- [ ] `arm/features.go` gains `//go:build arm.6` — `initFeatures()` only compiled for GOARM≥6.
- [ ] New `arm/features_v5.go` with `//go:build !arm.6` provides a no-op `func (cpu *CPU) initFeatures() {}`.
- [ ] The `if !cpu.NoVBAR { cpu.initFeatures() }` block in `arm/arm.go Init()` is removed; `initFeatures()` is called unconditionally (the build tag on `features.go`/`features_v5.go` handles it).
- [ ] `arm.CPU.Init()` reverts to zero arguments; `initVectorTable` derives the VBAR from `goos.RamStart` internally (restoring original behavior). `NUC980` SoC package updated to call `ARM.Init()`.
- [ ] `GOOS=tamago GOARCH=arm GOARM=7 go build ./...` produces zero errors (all existing boards compile).
- [ ] `GOOS=tamago GOARCH=arm GOARM=5 go build ./soc/nuvoton/nuc980/...` compiles cleanly.
- [ ] `GOOS=tamago GOARCH=arm GOARM=6 go build ./arm/...` compiles cleanly (ARM11 build-tag path).
- [ ] `arm.go` package doc comment updated to include `ARMv6 / ARM1176JZS` in the supported cores list.

**Verification:**
```bash
cd /home/mdr164/private/kyanite/tamago
GOOS=tamago GOARCH=arm GOARM=7 go build ./...
GOOS=tamago GOARCH=arm GOARM=6 go build ./arm/...
GOOS=tamago GOARCH=arm GOARM=5 go build ./soc/nuvoton/...
GOOS=tamago GOARCH=arm GOAM=5 go build ./board/nuvoton/...
```

**Dependencies:** None

**Files to change (tamago repo):**
- `arm/features.go` — add `//go:build arm.6` at top
- `arm/features_v5.go` — new file: `//go:build !arm.6`, no-op `initFeatures()`
- `arm/arm.go` — remove the `if !cpu.NoVBAR` guard; revert `Init(vbar uint32)` → `Init()`, use `uint32(goos.RamStart)` inside `initVectorTable`
- `soc/nuvoton/nuc980/nuc980.go` — change `ARM.Init(ramStart)` → `ARM.Init()`

**Size:** S

---

### Checkpoint 0

**Gate:** Both T0.1 and T0.2 complete. Run:
```bash
# tamago repo: all existing boards still compile
GOOS=tamago GOARCH=arm GOARM=7 go build ./...
# chiptool: generates correct output
cargo test -C /home/mdr164/private/kyanite/chiptool
```
Review output format from chiptool with a sample YAML before proceeding.

---

## Phase 1: AST2600 (Cortex-A7, ARMv7) — First Vertical Slice

> AST2600 is implemented first because it reuses the existing `arm` ARMv7 package with no
> new assembly, uses the standard `arm/gic` package, and uses the ARM generic timer
> already supported by `arm.InitGenericTimers`. This gives the fastest path to a working boot.

---

### Task 1.1: AST2600 chiptool YAML + generated register files

**Description:** Write the peripheral register YAML for AST2600 and run chiptool to generate the Go register accessor files that the SoC package will build on.

**Datasheets to read:**
- `ast2600v16.md` §2 Memory Map — confirm all base addresses
- §SCU (System Control Unit) — protection key, clock-gate bits, silicon-rev register at `+0x014`
- §UART — 16550 register layout (should match standard 16550)
- §Timer Controller — confirm FTTMR010 register layout at `0x1e782000`
- §HACE/TRNG — TRNG control/data registers (TRNG is part of HACE block or standalone)

**YAML file to create:** `soc/aspeed/ast2600/regs.yaml`

Minimum peripheral blocks required for boot:
```yaml
# block/Scu — System Control Unit
#   protection_key   +0x000  WO  unlock with 0x1688a8a8 / lock with 0x1
#   chip_id2         +0x014  RO  silicon revision
#   hw_strap1        +0x500  RO  CPU freq selector bits [10:8]
#   clkgate_ctrl1    +0x080  RW  peripheral clock gates
#   clkgate_clr1     +0x084  WO  write 1s to enable clocks (clear gate)

# block/Uart  (16550-compatible, base 0x1e784000)
#   thr/rbr   +0x00  tx/rx
#   ier       +0x04
#   fcr/iir   +0x08
#   lcr       +0x0c
#   mcr       +0x10
#   lsr       +0x14
#   msr       +0x18
#   dll       +0x00 (DLAB=1)
#   dlm       +0x04 (DLAB=1)

# block/Timer  (FTTMR010, base 0x1e782000, 8 timers)
#   timer[n].status    +0x00 + n*0x10
#   timer[n].reload    +0x04 + n*0x10
#   timer[n].match1    +0x08 + n*0x10
#   timer[n].match2    +0x0c + n*0x10
#   ctrl1              +0x30
#   ctrl2              +0x34

# block/Trng  — confirm register offsets from datasheet
# device/Ast2600 — all peripheral instances + IRQ numbers
```

**Acceptance criteria:**
- [ ] `soc/aspeed/ast2600/regs.yaml` exists and passes `chiptool validate`.
- [ ] Running `chiptool generate --language go --package ast2600 --reg-import github.com/usbarmory/tamago/internal/reg --output soc/aspeed/ast2600/ soc/aspeed/ast2600/regs.yaml` produces: `scu.gen.go`, `uart.gen.go`, `timer.gen.go`, `trng.gen.go`, `device.gen.go` — all with `package ast2600`.
- [ ] Generated files compile: `GOOS=tamago GOARCH=arm GOARM=7 go build ./soc/aspeed/ast2600/` (will need stub Go files for missing symbols, but generated files themselves must have no syntax errors).
- [ ] `device.gen.go` contains `SCU_BASE = 0x1e6e2000`, `UART5_BASE = 0x1e784000`, `TIMER_BASE = 0x1e782000`, `GIC_BASE = 0x40460000`.

**Verification:**
```bash
cd /home/mdr164/private/kyanite/chiptool
./target/debug/chiptool generate --language go \
  --package ast2600 \
  --reg-import github.com/usbarmory/tamago/internal/reg \
  --output ../tamago/soc/aspeed/ast2600/ \
  ../tamago/soc/aspeed/ast2600/regs.yaml
GOOS=tamago GOARCH=arm GOARM=7 go vet ./soc/aspeed/ast2600/...
```

**Dependencies:** T0.1 (chiptool CLI), T0.2 (arm package must compile)

**Files created:**
- `soc/aspeed/ast2600/regs.yaml`
- `soc/aspeed/ast2600/scu.gen.go` (generated)
- `soc/aspeed/ast2600/uart.gen.go` (generated)
- `soc/aspeed/ast2600/timer.gen.go` (generated)
- `soc/aspeed/ast2600/trng.gen.go` (generated)
- `soc/aspeed/ast2600/device.gen.go` (generated)

**Size:** S

---

### Task 1.2: `soc/aspeed/ast2600` SoC package

**Description:** Implement the AST2600 SoC package: ARM Cortex-A7 core initialization, GIC-400 interrupt controller, ARM generic timer for nanotime, TRNG for RNG, UART5 for console, and SCU for clock gates and silicon revision.

**Architecture notes:**
- Cortex-A7 on AST2600 runs in SYS mode at boot (no HYP mode — BMC SoC does not boot at EL2).
- ARM generic timer CNTFRQ must be set from `SCU_HW_STRAP1` bits [10:8] before `ARM.InitGenericTimers()` is called. Mapping: `0b000`/`0b010` → 1,200,000,000 Hz; `0b001`/`0b011` → 1,600,000,000 Hz; `0b100+` → 800,000,000 Hz.
- GIC: `Base = GIC_BASE` (0x40460000); `secure=false`, `fiqen=false` for initial boot.
- TamaGo is the **only user software** on the SoC. The ASPEED hardware boot ROM (internal mask ROM, not user-flashable) initializes SDRAM and loads the TamaGo FMC image from SPI flash into SDRAM before jumping to the TamaGo entry point. No U-Boot or other bootloader is involved. A TamaGo FMC image packager tool is required (see T1.3 notes).
- ⚠️ **AST2400/AST2500 note:** The ASPEED boot ROM remaps SDRAM to `0x00000000` after init. Confirm from datasheets whether TamaGo should use `goos.RamStart=0x0` (remapped) or the canonical SDRAM base. See Risk R9.
- AST2600 also has a legacy SWVIC at `0x1e6c0000` — ignore it; use GIC only.

**Files to create:**

`soc/aspeed/ast2600/ast2600.go` — package doc + peripheral instances referencing constants from `device.gen.go`:
```go
package ast2600

import (
    "github.com/usbarmory/tamago/arm"
    "github.com/usbarmory/tamago/arm/gic"
)

var (
    ARM  = &arm.CPU{}
    GIC  = &gic.GIC{Base: GIC_BASE}
    // UART5 and TIMER1 use the generated Uart/Timer struct types
    UART5  = &Uart{Base: UART5_BASE}
    TIMER1 = &Timer{Base: TIMER_BASE}
    TRNG   = &Trng{Base: TRNG_BASE}
)
```

`soc/aspeed/ast2600/init.go` — `Init()` and `init()`:
- `Init()`: unlock SCU, ARM.Init(), GIC.Init(false,false), ARM.EnableSMP(), ARM.InitMMU(), ARM.EnableCache(), initTimers(), UART5.Init(), GIC enable scheduler-tick IRQ
- `init()`: read SCU silicon rev, set exported `SiliconRevision` var

`soc/aspeed/ast2600/timer.go`:
- `initTimers()`: read HW_STRAP1, compute CNTFRQ, call `ARM.InitGenericTimers(0, freq)`
- `nanotime()` linkname to `runtime/goos.Nanotime` → `ARM.GetTime()`
- `initRNG()` linkname to `runtime/goos.InitRNG`
- `getRandomData()` linkname to `runtime/goos.GetRandomData` (via TRNG)

`soc/aspeed/ast2600/rng.go` — TRNG driver (register reads from `trng.gen.go` struct)

`soc/aspeed/ast2600/mem.go`:
```go
//go:build !linkramstart
//go:linkname ramStart runtime/goos.RamStart
var ramStart uint32 = 0x80000000

//go:linkname ramStackOffset runtime/goos.RamStackOffset
var ramStackOffset uint32 = 0x100000 // 1 MB
```

`soc/aspeed/ast2600/clock.go` — SCU clock gate helpers: `EnableUARTClock()`, `EnableTimerClock()`

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=7 go build ./soc/aspeed/ast2600/` compiles with zero errors.
- [ ] `GOOS=tamago go vet ./soc/aspeed/ast2600/` produces no warnings.
- [ ] Package-level `go doc ./soc/aspeed/ast2600` shows correct doc comment referencing AST2600 datasheet.
- [ ] `SiliconRevision` is exported and readable.
- [ ] `ARM`, `GIC`, `UART5`, `TIMER1`, `TRNG` are exported package-level vars.

**Dependencies:** T0.2, T1.1

**Files created:** `ast2600.go`, `init.go`, `timer.go`, `rng.go`, `mem.go`, `clock.go`

**Size:** M

---

### Task 1.3: `board/aspeed/ast2600evb` board package

**Description:** Implement the AST2600EVB reference board package: `cpuinit.s` for stack initialization and entry, `Hwinit1` linkname, `Printk` linkname via UART5, and RAM size/start.

**Notes:**
- AST2600 has VBAR (Cortex-A7), so `cpuinit.s` does NOT need to install exception stubs at `0x0`. It only needs to set the stack pointer, ensure SVC mode, clear BSS, and branch to `_rt0_tamago_start`. This is nearly identical to `arm/init.s` but gated by `linkcpuinit`.
- The AST2600 boot ROM always enters in SVC mode. No HYP mode on this SoC. Keep the HYP-mode ERET check from `arm/init.s` for safety but it will never trigger.
- Default RAM size: 512 MB (`0x20000000`). Override with `-tags linkramsize` + custom `mem.go`.

**Files to create:**

`board/aspeed/ast2600evb/cpuinit.s`:
```asm
//go:build linkcpuinit
// Sets stack from RamStart+RamSize-RamStackOffset.
// Handles optional HYP→SVC eret (same as arm/init.s).
// Clears BSS. Branches to _rt0_tamago_start.
```

`board/aspeed/ast2600evb/ast2600evb.go`:
```go
package ast2600evb

//go:linkname Init runtime/goos.Hwinit1
func Init() {
    ast2600.Init()
    ast2600.UART5.Init()
}
```

`board/aspeed/ast2600evb/console.go` (`//go:build !linkprintk`):
```go
//go:linkname printk runtime/goos.Printk
func printk(c byte) {
    if c == '\n' { ast2600.UART5.Tx('\r') }
    ast2600.UART5.Tx(c)
}
```

`board/aspeed/ast2600evb/mem.go` (`//go:build !linkramsize`):
```go
//go:linkname ramSize runtime/goos.RamSize
var ramSize uint32 = 0x20000000 // 512 MB
```

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=7 go build -tags linkcpuinit ./board/aspeed/ast2600evb/` compiles.
- [ ] A minimal `cmd/ast2600test/main.go` that blank-imports the board package builds end-to-end:
  ```bash
  GOOS=tamago GOARCH=arm GOARM=7 go tool tamago build \
    -tags linkcpuinit \
    -ldflags "-T 0x80010000 -R 0x1000" \
    ./cmd/ast2600test/
  ```
  produces a valid ELF (`file ast2600test` reports ARM ELF).

**Dependencies:** T0.2, T1.2

**Files created:** `cpuinit.s`, `ast2600evb.go`, `console.go`, `mem.go`

**Size:** S

---

### Task 1.4: AST2600 QEMU smoke test

**Description:** Boot the minimal test binary under QEMU and verify the standardized boot banner appears on UART within 10 seconds.

**Command:**
```bash
qemu-system-arm \
  -machine ast2600-evb \
  -cpu cortex-a7 \
  -m 512M \
  -nographic \
  -serial null \
  -serial stdio \
  -kernel ast2600test.elf \
  -no-reboot
```

**Expected output (must contain):**
```
Hello from TamaGo on ASPEED AST2600!
Runtime :
Board   : AST2600EVB
CPU     : Cortex-A7
RAM     :
```

**Acceptance criteria:**
- [ ] QEMU exits (or is killed) within 10 seconds of launch without a kernel panic.
- [ ] UART output contains the boot banner lines above.
- [ ] No `panic:` or `exception: vector` lines appear.
- [ ] A `Makefile` target `make qemu-ast2600` is added to the repo root (or a `board/aspeed/ast2600evb/Makefile`).

**Verification:**
```bash
make qemu-ast2600 2>&1 | tee /tmp/ast2600-qemu.log
grep -q "Hello from TamaGo on ASPEED AST2600" /tmp/ast2600-qemu.log
```

**Dependencies:** T1.3

**Files created/modified:** `Makefile` or `board/aspeed/ast2600evb/Makefile`

**Size:** XS

---

### Task 1.5: AST2600 hardware verification

**Description:** Boot on real AST2600 EVB hardware. TamaGo runs as the sole user software; the ASPEED hardware boot ROM loads the FMC image from SPI flash. Alternatively, JTAG can inject the binary directly for development.

**Loading method (FMC image → SPI flash):**
Use the TamaGo FMC image packager (built in T1.3) to pack the ELF into an ASPEED FMC-compatible image, then flash it to the SPI chip with a programmer or via the boot ROM's SDP mode.

**Loading method (JTAG / OpenOCD, development only):**
```
load_image ast2600test.elf
resume 0x80010000
```

**Acceptance criteria:**
- [ ] Serial capture shows boot banner including a decoded silicon revision string (e.g. `AST2600 A3`).
- [ ] Goroutine scheduler tick observable (at least 5 "tick N" lines if the test app prints them).
- [ ] Boot log committed to `board/aspeed/ast2600evb/README.md` under a `## Hardware Boot Log` section.

**Dependencies:** T1.4 (QEMU must pass first as sanity check)

**Files created/modified:** `board/aspeed/ast2600evb/README.md`

**Size:** XS

---

### Checkpoint 1

**Gate:** T1.1–T1.5 complete. QEMU boot passes. Hardware boot log committed.
```bash
GOOS=tamago GOARCH=arm GOARM=7 go build ./...   # zero errors
make qemu-ast2600                                 # banner appears, no panic
```
Review silicon revision decode, timer frequency selection, and GIC base address
with the AST2600 datasheet before proceeding. **This is the point to raise any
upstream `tamago` concerns (GIC Base naming, goos linkname conventions, etc.).**

---

## Phase 2: Shared Peripheral Drivers + AST2500 (ARM1176, ARMv6)

> The shared `soc/aspeed/uart/`, `soc/aspeed/timer/`, and `soc/aspeed/intc/`
> packages are introduced here because AST2400 and AST2500 share identical peripheral
> bases. Extracting drivers once avoids duplicating 200+ lines per SoC.

---

### Task 2.1 [P]: `soc/aspeed/uart` — shared 16550 UART driver

**Description:** Implement a reusable 16550-compatible UART driver used by AST2400, AST2500, and AST2600. Mirrors the existing `soc/nxp/uart` pattern.

**Notes:**
- The `Uart` struct generated by chiptool (in each SoC package) provides `ReadLsr()`, `WriteThr()`, etc. The shared driver can use `internal/reg` directly with a `Base uint32` field — no dependency on the generated struct type.
- 16550 register offsets are fixed: THR/RBR=+0, IER=+4, FCR/IIR=+8, LCR=+C, MCR=+10, LSR=+14, DLL=+0(DLAB), DLM=+4(DLAB).
- Baud rate formula: `DLL:DLM = (clk / (16 * baud))`. For 115200 baud at 24 MHz UART clock: divisor = `24000000/(16*115200)` = 13 (0x0D).  ⚠️ Confirm UART clock source from datasheets — may differ per SoC.

**Files to create:** `soc/aspeed/uart/uart.go`
```go
package uart

type UART struct {
    Index int
    Base  uint32
    Clock func() uint  // returns UART input clock in Hz
}

func (hw *UART) Init() { ... }  // set LCR, FCR, baud divisor
func (hw *UART) Tx(c byte) { ... }  // wait LSR_THRE, write THR
func (hw *UART) Write(b []byte) { ... }
```

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=7 go build ./soc/aspeed/uart/` compiles.
- [ ] `GOOS=tamago GOARCH=arm GOARM=5 go build ./soc/aspeed/uart/` compiles.
- [ ] `GOOS=tamago go test -tags user_linux ./soc/aspeed/uart/` — at minimum, `go vet` passes (unit test stubs acceptable if hardware register access is abstracted).

**Dependencies:** T0.2

**Files:** `soc/aspeed/uart/uart.go`

**Size:** XS

---

### Task 2.2 [P]: `soc/aspeed/timer` — shared FTTMR010 MMIO timer driver

**Description:** Implement the FTTMR010 free-running and periodic timer driver used by AST2400 and AST2500. (AST2600/AST2700 use the ARM generic timer instead.)

**Register layout** (from `u-boot/arch/arm/include/asm/arch-aspeed/timer.h` and AST2500 datasheet):
```
Base = 0x1e782000
Timer N (N=1..8):
  +0x00 + (N-1)*0x10  status (current count, counts down)
  +0x04 + (N-1)*0x10  reload_val
  +0x08 + (N-1)*0x10  match1
  +0x0c + (N-1)*0x10  match2
ctrl1  +0x30  (4 bits per timer: [EN|1MHZ|OVFINTR|PWM])
ctrl2  +0x34
```

**Files to create:** `soc/aspeed/timer/timer.go`, `soc/aspeed/timer/timer.s`

```go
package timer

type Timer struct {
    Index int    // 1-8
    Base  uint32 // 0x1e782000
}

func (hw *Timer) InitFreeRunning() { ... }   // 1 MHz, continuous, no interrupt
func (hw *Timer) InitPeriodic(us uint32) { ... } // overflow interrupt at us interval
func (hw *Timer) Read() uint32 { ... }        // current counter value (assembly)
func (hw *Timer) AckInterrupt() { ... }       // clear overflow interrupt
```

`timer.s`: assembly `readTimer(base uint32) uint32` loading the status register.

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=5 go build ./soc/aspeed/timer/` compiles.
- [ ] `GOOS=tamago GOARCH=arm GOARM=6 go build ./soc/aspeed/timer/` compiles.
- [ ] `GOOS=tamago GOARCH=arm GOARM=7 go build ./soc/aspeed/timer/` compiles.

**Dependencies:** T0.2

**Files:** `soc/aspeed/timer/timer.go`, `soc/aspeed/timer/timer.s`

**Size:** XS

---

### Task 2.3 [P]: `soc/aspeed/intc` — shared ASPEED VIC interrupt controller

**Description:** Implement the ASPEED VIC (ARM PL192-compatible) interrupt controller driver used by AST2400 and AST2500.

**VIC register map** (from AST2400/AST2500 datasheets, base `0x1e6c0000`):
- Read the `ast2400v14.md` and `ast2500v18.md` §VIC register table for exact offsets before implementing.
- Expected offsets (ARM PL192 standard): IRQStatus=+0x00, FIQStatus=+0x04, RawIntr=+0x08, IntSelect=+0x0c, IntEnable=+0x10, IntEnClear=+0x14, SoftInt=+0x18, SoftIntClear=+0x1c, VectAddr=+0xf00 (current), DefVectAddr=+0xf04.
- ⚠️ The u-boot DTS places the VIC at `0x1e6c0080`, not `0x1e6c0000`. The u-boot `platform.S` uses absolute addresses `0x1e6c0038` (VIC timer IRQ clear) and `0x1e6c0090`. Cross-check with datasheet to resolve this discrepancy before coding.

**Files to create:** `soc/aspeed/intc/intc.go`
```go
package intc

type INTC struct {
    Base uint32
}

func (hw *INTC) Init() { ... }           // disable all IRQs
func (hw *INTC) EnableIRQ(irq int) { ... }
func (hw *INTC) DisableIRQ(irq int) { ... }
func (hw *INTC) CurrentIRQ() int { ... } // reads VICVectAddr or status
func (hw *INTC) EOI() { ... }            // End-of-interrupt
func (hw *INTC) SoftwareInterrupt(irq int) { ... }
```

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=5 go build ./soc/aspeed/intc/` compiles.
- [ ] `GOOS=tamago GOARCH=arm GOARM=6 go build ./soc/aspeed/intc/` compiles.
- [ ] VIC base address and all register offsets include a comment citing the datasheet page number.
- [ ] `intc.go` has a `// ⚠️ Verify base address: DTS uses 0x1e6c0080, datasheet may differ` note resolved to confirmed offset.

**Dependencies:** T0.2

**Files:** `soc/aspeed/intc/intc.go`

**Size:** S

---

### Task 2.4: AST2500 chiptool YAML + generated register files

**Description:** Write AST2500 peripheral YAML and generate register accessor files. AST2500 shares identical peripheral addresses with AST2400 (same UART, timer, VIC, SCU bases) but differs in silicon revision encoding and clock tree details.

**YAML file to create:** `soc/aspeed/ast2500/regs.yaml`

Required blocks:
- `block/Scu` — protection_key, rev_id (at `+0x07c`), hw_strap (at `+0x070`), clk_stop_ctrl1 (at `+0x0c`)
- `block/Uart` — identical 16550 layout as AST2600 YAML (can reference via YAML anchors or copy)
- `block/Timer` — FTTMR010 layout identical to AST2600 YAML
- `block/Vic` — VIC register map (offsets confirmed in T2.3)
- `device/Ast2500` — instances: `UART5_BASE=0x1e784000`, `TIMER_BASE=0x1e782000`, `VIC_BASE=0x1e6c0000`, `SCU_BASE=0x1e6e2000` + IRQ numbers

**Acceptance criteria:**
- [ ] `chiptool validate soc/aspeed/ast2500/regs.yaml` passes.
- [ ] Generated files produced in `soc/aspeed/ast2500/`: `scu.gen.go`, `uart.gen.go`, `timer.gen.go`, `vic.gen.go`, `device.gen.go`.
- [ ] `device.gen.go` contains correct IRQ constant for Timer1 overflow interrupt (read from ast2500v18.md VIC interrupt table).
- [ ] `GOOS=tamago GOARCH=arm GOARM=6 go vet ./soc/aspeed/ast2500/` passes on generated files alone.

**Dependencies:** T0.1, T2.3 (VIC register offsets confirmed)

**Files:** `soc/aspeed/ast2500/regs.yaml` + 5 generated `.gen.go` files

**Size:** S

---

### Task 2.5: `soc/aspeed/ast2500` SoC package

**Description:** Implement the AST2500 SoC package: ARM1176 core init with `NoVBAR=true`, ASPEED VIC interrupt controller, FTTMR010 timer for nanotime and scheduler tick, UART5 for console, software RNG.

**Architecture notes:**
- ARM1176JZS does NOT have a WFI opcode by default — it uses `MCR p15,0,r0,c7,c0,4` (same as ARM926). The `irq_v5.s` WFI implementation (via `!arm.6`) is used. ⚠️ **Actually**: ARM1176JZF-S (as in AST2500) IS ARMv6K which DOES have the `WFI` instruction. The `//go:build arm.6` branch in `irq.s` uses `WFI` directly. This is correct.
- `NoVBAR=true`: no VBAR register on ARM1176. Exception stubs installed at `0x0` by `cpuinit.s`.
- The VIC-based scheduler tick is an overflow interrupt from Timer1 at 10ms period (Timer IRQ period configurable from SoC package constant).
- Software RNG: use `internal/rng` package with a seed derived from the free-running timer counter at init time. No TRNG on AST2500.
- `initFeatures()` now runs on ARM1176 (GOARM=6, build tag restored in T0.2). Feature detection will correctly identify TrustZone and other ARM1176 capabilities.

**Files to create (same structure as ast2600):**
- `ast2500.go` — ARM, INTC, UART5, TIMER1 instances; imports `soc/aspeed/uart`, `soc/aspeed/timer`, `soc/aspeed/intc`
- `init.go` — `Init()`: SCU unlock, `ARM.Init()`, `INTC.Init()`, `initTimers()`, `UART5.Init()`, enable Timer IRQ
- `timer.go` — `initTimers()` configuring TIMER1 as 1 MHz free-running + TIMER2 as 10ms tick; `nanotime()` linkname
- `rng.go` — `initRNG()` + `getRandomData()` using software PRNG seeded from timer
- `mem.go` — `ramStart = 0x80000000`, `ramStackOffset = 0x100000`
- `clock.go` — SCU clock gate helpers reading ast2500 clk_stop_ctrl1

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=6 go build ./soc/aspeed/ast2500/` compiles.
- [ ] `GOOS=tamago go vet ./soc/aspeed/ast2500/` passes.
- [ ] ARM926EJ-S comparison: both `soc/aspeed/ast2400/` and `soc/aspeed/ast2500/` can be built simultaneously without conflict.

**Dependencies:** T0.2, T2.1, T2.2, T2.3, T2.4

**Files:** `ast2500.go`, `init.go`, `timer.go`, `rng.go`, `mem.go`, `clock.go`

**Size:** M

---

### Task 2.6: `board/aspeed/ast2500evb` board package

**Description:** Implement the AST2500EVB board package with ARM1176-specific `cpuinit.s` that installs exception stubs at `0x0`, calls SCU EarlyInit, configures UART5 pin mux, then enters the Go runtime.

**`cpuinit.s` specifics (GOARM=6, `//go:build linkcpuinit`):**
1. Write `0xEAFFFFFE` (infinite loop branch) to `0x00000000..0x0000001c` (8 × 4-byte exception vector slots).
2. Call `ast2500.EarlyInit` (pure assembly function that unlocks SCU + enables UART/timer clocks).
3. Configure UART5 pin mux: read AST2500 datasheet for SCU pinmux register and UART5 mux bits.
4. Clear BSS (same pattern as NUC980 `cpuinit.s`).
5. Set stack: `RamStart + RamSize - RamStackOffset`.
6. Enter SVC mode (ARM1176 boots in SVC; no HYP mode).
7. Branch to `_rt0_tamago_start`.

`earlyinit.s` in `soc/aspeed/ast2500/` (pure assembly, no Go runtime):
- Unlock SCU: `MOV 0x1e6e2000 → write 0x1688a8a8`
- Enable UART5 clock gate in SCU clock-stop register (bit from ast2500v18.md)
- Enable Timer1/2 clock gates

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=6 go build -tags linkcpuinit ./board/aspeed/ast2500evb/` compiles.
- [ ] Minimal test binary builds:
  ```bash
  GOOS=tamago GOARCH=arm GOARM=6 go tool tamago build \
    -tags linkcpuinit \
    -ldflags "-T 0x80010000 -R 0x1000" \
    ./cmd/ast2500test/
  ```
- [ ] `file ast2500test.elf` reports ARM ELF (not Thumb).

**Dependencies:** T2.5

**Files:** `cpuinit.s`, `ast2500evb.go`, `console.go`, `mem.go`; `soc/aspeed/ast2500/earlyinit.s`

**Size:** S

---

### Task 2.7: AST2500 QEMU and hardware verification

**QEMU:**
```bash
qemu-system-arm \
  -machine ast2500-evb \
  -cpu arm1176 \
  -m 512M \
  -nographic \
  -serial null \
  -serial stdio \
  -kernel ast2500test.elf
```

**Acceptance criteria:**
- [ ] QEMU boots and prints banner containing `AST2500` and `ARM1176`.
- [ ] No panic or undefined exception.
- [ ] Hardware boot log captured and committed to `board/aspeed/ast2500evb/README.md`.
- [ ] Banner includes decoded silicon revision from SCU`+0x07C`.

**Dependencies:** T2.6

**Size:** XS

---

### Checkpoint 2

**Gate:** T2.1–T2.7 complete. Both AST2600 and AST2500 boot on QEMU and hardware.
```bash
GOOS=tamago GOARCH=arm GOARM=7 go build ./...
GOOS=tamago GOARCH=arm GOARM=6 go build ./...
make qemu-ast2600 qemu-ast2500
```
Confirm VIC base address resolution (0x1e6c0000 vs 0x1e6c0080) is resolved and documented.

---

## Phase 3: AST2400 (ARM926EJ-S, ARMv5)

> AST2400 is implemented after AST2500 because it reuses all three shared drivers
> (uart, timer, intc) with **identical base addresses**. The only difference is
> DRAM base (0x40000000 vs 0x80000000) and GOARM=5.

---

### Task 3.1: AST2400 chiptool YAML + generated register files

**Description:** Write AST2400 peripheral YAML. Peripheral register layouts are identical to AST2500 (same UART, Timer, VIC, SCU bases and register formats); only the `device/Ast2400` section differs (silicon rev value, chip-specific IRQ assignments).

**YAML file:** `soc/aspeed/ast2400/regs.yaml`

The `block/` entries for `Scu`, `Uart`, `Timer`, `Vic` can be shared/referenced from `ast2500/regs.yaml` using chiptool's transform/merge mechanism, or simply duplicated. Confirm with the chiptool YAML format whether cross-file references are supported.

Key differences from ast2500 YAML:
- `device/Ast2400` base address: `DRAM_BASE = 0x40000000` (instead of `0x80000000`)
- SCU silicon rev at `+0x07C`, but reset value is `0x02000303` for AST2400 A1 (vs AST2500)
- IRQ numbers for timer: verify from `ast2400v14.md` VIC IRQ table

**Acceptance criteria:**
- [ ] Generated `device.gen.go` contains `DRAM_BASE = 0x40000000`.
- [ ] `GOOS=tamago GOARCH=arm GOARM=5 go vet ./soc/aspeed/ast2400/` passes on generated files.

**Dependencies:** T0.1, T2.4 (AST2500 YAML establishes the block format)

**Files:** `soc/aspeed/ast2400/regs.yaml` + 5 generated `.gen.go` files

**Size:** XS

---

### Task 3.2: `soc/aspeed/ast2400` SoC package

**Description:** Implement the AST2400 SoC package: ARM926EJ-S with `NoVBAR=true`, ASPEED VIC, FTTMR010 timer, UART5, software RNG. Structure mirrors `soc/aspeed/ast2500/` with these differences:
- `GOARM=5`, ARMv5 (soft-float, no CPSIE/CPSID, no WFI opcode — use `irq_v5.s`)
- `ramStart = 0x40000000`
- Silicon revision from `ast2400v14.md`
- No TRNG, no TrustZone

**Notes:**
- Exception vectors at `0x00000000` — but wait: on AST2400, DRAM is at `0x40000000`. The exception vectors on ARM926 are at physical `0x00000000` which is the internal SRAM (`0x1e720000` alias at `0x0`), not DRAM. The `cpuinit.s` must write exception stubs to `0x00000000` (SRAM, not DRAM). Confirm this mapping from `ast2400v14.md` §Memory Map before implementing `cpuinit.s`.
- `ramStart = 0x40000000`, `ramStackOffset = 0x100000`

**Files:** `ast2400.go`, `init.go`, `timer.go`, `rng.go`, `mem.go`, `clock.go`, `earlyinit.s`

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm GOARM=5 go build ./soc/aspeed/ast2400/` compiles.
- [ ] `GOOS=tamago GOARCH=arm GOARM=5 go vet ./soc/aspeed/ast2400/` passes.

**Dependencies:** T0.2, T2.1, T2.2, T2.3, T3.1

**Files:** 6 Go files + 1 assembly file

**Size:** M

---

### Task 3.3: `board/aspeed/ast2400evb` board package

**Description:** Board package for the AST2400 EVB. Nearly identical to `ast2500evb` except:
- `GOARM=5` → uses `irq_v5.s` / `exception_v5.s` paths automatically
- `RamStart = 0x40000000`
- Exception stubs written to `0x00000000` (internal SRAM alias, not DRAM)
- `cpuinit.s` must not use CPSIE/CPSID or WFI instructions (ARMv5 constraint)
- Linker: `-T 0x40010000`

**Acceptance criteria:**
- [ ] Build succeeds: `GOOS=tamago GOARCH=arm GOARM=5 go tool tamago build -tags linkcpuinit -ldflags "-T 0x40010000 -R 0x1000" ./cmd/ast2400test/`
- [ ] `file ast2400test.elf` → ARM ELF.

**Dependencies:** T3.2

**Files:** `cpuinit.s`, `ast2400evb.go`, `console.go`, `mem.go`

**Size:** S

---

### Task 3.4: AST2400 QEMU and hardware verification

**QEMU:**
```bash
qemu-system-arm \
  -machine palmetto-bmc \
  -cpu arm926 \
  -m 128M \
  -nographic \
  -serial stdio \
  -kernel ast2400test.elf
```

**Acceptance criteria:**
- [ ] Banner contains `AST2400` and `ARM926EJ-S`.
- [ ] Hardware boot log committed to `board/aspeed/ast2400evb/README.md`.
- [ ] No ARMv5 illegal-instruction faults (would indicate CPSIE/WFI slipping through build tags).

**Dependencies:** T3.3

**Size:** XS

---

### Checkpoint 3

**Gate:** T3.1–T3.4 complete. All three 32-bit SoCs boot on QEMU and hardware.
```bash
for goarm in 5 6 7; do
  GOOS=tamago GOARCH=arm GOARM=$goarm go build ./... 2>&1 | grep -E "^(#|FAIL)"
done
make qemu-ast2400 qemu-ast2500 qemu-ast2600
```
All three must produce clean output and boot banners.

---

## Phase 4: AST2700 (Cortex-A35, AArch64)

> AST2700 is implemented last due to the significant architectural differences:
> 64-bit address space, DRAM above 4 GB, GICv3, different SCU/timer/UART bases,
> and SMP mailbox protocol for multi-core. This phase is single-core only.

---

### Task 4.1 [P]: AST2700 address space and linker validation

**Description:** Validate that the Go arm64 toolchain and TamaGo runtime correctly handle DRAM at `0x400000000` (above 4 GB). This is a prerequisite for any AST2700 work; a failed linker flag here would invalidate all address decisions.

**Steps:**
1. Write a trivial `cmd/ast2700probe/main.go` that imports only `unsafe` and prints `goos.RamStart`.
2. Attempt to build:
   ```bash
   GOOS=tamago GOARCH=arm64 go tool tamago build \
     -ldflags "-T 0x400010000 -R 0x1000" \
     ./cmd/ast2700probe/
   ```
3. Verify the ELF entry point is at `0x400010000`:
   ```bash
   readelf -h ast2700probe.elf | grep "Entry point"
   # Expected: Entry point address: 0x400010000
   ```
4. Verify `runtime/goos.RamStart` linkname works with `var ramStart uint = 0x400000000` (uint is 64-bit on arm64).

**Acceptance criteria:**
- [ ] `readelf -h` shows entry point `0x400010000`.
- [ ] No linker overflow or truncation errors.
- [ ] A note is added to this task (inline in the ROADMAP or in a commit message) confirming the result.

**Dependencies:** T0.2

**Files:** `cmd/ast2700probe/main.go` (temporary, can be deleted after confirmation)

**Size:** XS

---

### Task 4.2: AST2700 chiptool YAML + generated register files

**Description:** Write AST2700 peripheral YAML. AST2700 has a fundamentally different address map from AST2400/2500/2600 (all peripherals on SoC0 at `0x12cXXXXX` or SoC1 at `0x14cXXXXX`).

**Key addresses** (from `ast2700v03.md` §Memory Map and u-boot `scu_ast2700.h`):
- SCU0 base: `0x12c02000` (CPU domain)
- UART4 (SoC0, console): `0x12c1a000`
- Timer0..7: `0x12c10000` + N×`0x40` stride (stride confirmed from datasheet)
- GICv3: base addresses from `ast2700v03.md` §GIC (GICD, GICR, GICC)
- ⚠️ Confirm all addresses against `ast2700v03.md` §2 Memory Map before writing YAML

**YAML file:** `soc/aspeed/ast2700/regs.yaml`
Required blocks: `Scu0`, `Uart` (16550 compatible), `Timer`, `Gic` (GICv3 GICD/GICR)
Required device: `Ast2700` with `DRAM_BASE=0x400000000`, SCU0/UART4/TIMER0 bases, GIC addresses, IRQ numbers

**Acceptance criteria:**
- [ ] `chiptool validate` passes.
- [ ] `device.gen.go` contains `DRAM_BASE` as a `uint64` constant (not `uint32` — the value overflows 32 bits).
- [ ] `GOOS=tamago GOARCH=arm64 go vet ./soc/aspeed/ast2700/` passes on generated files.

**Note on `DRAM_BASE` type:** The chiptool `device.rs` currently generates `0x{:x}` untyped integer literals. For `0x400000000`, Go will infer `uint64`. Confirm this does not cause a type mismatch with `var ramStart uint`. If it does, the chiptool device template may need a `uint` cast for DRAM_BASE constants.

**Dependencies:** T0.1, T4.1

**Files:** `soc/aspeed/ast2700/regs.yaml` + generated `.gen.go` files

**Size:** S

---

### Task 4.3: `soc/aspeed/ast2700` SoC package

**Description:** Implement the AST2700 SoC package for AArch64. Uses the `arm64` and `arm64/gic` packages from tamago. ARM generic timer (v8) for nanotime.

**Architecture notes:**
- `GOARCH=arm64`, no `arm` package dependency.
- Import `github.com/usbarmory/tamago/arm64` (not `arm`).
- Import `github.com/usbarmory/tamago/arm64/gic` for GICv3.
- ARM generic timer v8 (`CNTPCT_EL0`) — confirm CNTFRQ from AST2700 SCU0 clock registers (25 MHz crystal input; PLL-derived CPU clock). Read `ast2700v03.md` §SCU Clock Control before implementing.
- UART4 SoC0 at `0x12c1a000` — 16550 compatible, reuse `soc/aspeed/uart` driver.
- Hardware TRNG available on AST2700 — confirm register base from datasheet.
- Single-core only (CPU0): the SMP mailbox protocol (`SCU_CPU_SMP_READY = 0x12c02780`, magic `0xbabecafe`) is visible in u-boot `lowlevel_init.S`. Non-CPU0 cores will spin in the mailbox WFE loop. Do not start them.
- `ramStart uint = 0x400000000`, `ramStackOffset uint = 0x100000`.

**Files:** `ast2700.go`, `init.go`, `timer.go`, `rng.go`, `mem.go`, `clock.go`

**Acceptance criteria:**
- [ ] `GOOS=tamago GOARCH=arm64 go build ./soc/aspeed/ast2700/` compiles.
- [ ] `GOOS=tamago GOARCH=arm64 go vet ./soc/aspeed/ast2700/` passes.
- [ ] No `arm` package imported (AArch64 only).

**Dependencies:** T0.2, T2.1 (uart driver), T4.1, T4.2

**Files:** 6 Go files

**Size:** M

---

### Task 4.4: `board/aspeed/ast2700evb` board package

**Description:** AArch64 board package for the AST2700 EVB. The `cpuinit.s` must be AArch64 assembly.

**`cpuinit.s` specifics (AArch64, `//go:build linkcpuinit`):**
1. Read `MPIDR_EL1`; if not CPU0, spin in WFE loop (check SMP mailbox at `SCU_CPU_SMP_READY`).
2. Set `SP_EL1` (stack pointer) from `RamStart + RamSize - RamStackOffset`.
3. Call `ast2700.EarlyInit` (assembly: unlock SCU0, enable UART4 clock).
4. Configure UART4 SoC0 pin mux (read ast2700v03.md §SCU0 pinmux for UART4 bits).
5. Clear BSS.
6. Branch to `_rt0_tamago_start`.

**Linker flag:** `-T 0x400010000`

**Acceptance criteria:**
- [ ] Build succeeds: `GOOS=tamago GOARCH=arm64 go tool tamago build -tags linkcpuinit -ldflags "-T 0x400010000 -R 0x1000" ./cmd/ast2700test/`
- [ ] `readelf -h ast2700test.elf` entry point = `0x400010000`.
- [ ] `file ast2700test.elf` → `ARM aarch64`.

**Dependencies:** T4.3

**Files:** `cpuinit.s`, `ast2700evb.go`, `console.go`, `mem.go`; `soc/aspeed/ast2700/earlyinit.s`

**Size:** S

---

### Task 4.5: AST2700 hardware verification

**Description:** Boot on real AST2700 EVB hardware. QEMU AST2700 machine support is experimental; hardware testing is the primary gate.

**Loading:**
```
# Via JTAG / OpenOCD (development):
load_image ast2700test.elf
resume 0x400010000

# Via FMC image (production):
# Build FMC image with the packager tool, flash to SPI
```

**Acceptance criteria:**
- [ ] Serial output shows `AST2700` banner with `Cortex-A35`.
- [ ] `RAM: 16 GB @ 0x400000000` (or whatever the EVB has installed).
- [ ] Silicon revision decoded from SCU0`+0x000` matches `0x06000003` (AST2700 A0).
- [ ] Boot log committed to `board/aspeed/ast2700evb/README.md`.
- [ ] If QEMU support is confirmed working, a `make qemu-ast2700` target is added.

**Dependencies:** T4.4

**Size:** XS

---

### Final Checkpoint

**Gate:** All four SoCs boot on hardware. All four QEMU smoke tests pass (AST2700 QEMU gated separately). Full build matrix clean:
```bash
GOOS=tamago GOARCH=arm GOARM=5 go build ./...
GOOS=tamago GOARCH=arm GOARM=6 go build ./...
GOOS=tamago GOARCH=arm GOARM=7 go build ./...
GOOS=tamago GOARCH=arm64     go build ./...
```
All four README boot logs committed. Ready for upstream PR to `usbarmory/tamago`.

---

## Risks and Open Questions

### Risks

| ID | Risk | Impact | Mitigation |
|---|---|---|---|
| R1 | AST2400 exception vectors at `0x0` mapped to internal SRAM, not DRAM (DRAM at `0x40000000`) | **High** — boot will fault if stubs are written to DRAM instead | Read `ast2400v14.md` §Memory Map before T3.3. Confirm SRAM alias at `0x0` exists. If not, DRAM-to-0 remap via SCU may be needed. |
| R2 | ASPEED VIC base address discrepancy (DTS: `0x1e6c0080`, u-boot platform.S uses absolute `0x1e6c0038/0x1e6c0090`) | **High** — wrong base silently fails (no IRQ delivery) | Resolve in T2.3 before writing any code: read `ast2400v14.md` §VIC register map for authoritative offsets. |
| R3 | AST2700 DRAM at `0x400000000` — Go arm64 linker may reject `-T 0x400010000` | **High** — entire AST2700 phase blocked | Validated in T4.1 before any AST2700 code is written. If rejected, investigate `GOARM64` or boot at a SRAM alias first. |
| R4 | AST2600 GIC base computed as `0x40460000` from DTS; actual periphbase may differ | **Med** — GIC init silently ignores interrupts | Confirm from `ast2600v16.md` §GIC section. Use `GICD_CTLR` read-back (should return non-zero) as sanity check in `Init()`. |
| R5 | AST2600 UART clock source and baud divisor | **Med** — garbled console output | Read `ast2600v16.md` §SCU clock tree and §UART clock source register before T1.2. U-Boot `arch/arm/mach-aspeed/ast2600/clk.c` confirms the clock. |
| R6 | NUC980 PR #66 not yet merged upstream; `aspeed` branch may diverge from `development` | **Med** — rebase work required | Track PR status. If upstream makes structural changes to `arm.CPU`, apply them before T1.2. |
| R7 | `chiptool` Go backend generates `uint32` for all addresses; `0x400000000` overflows | **Med** — build error on AST2700 device.gen.go | Address in T4.2. May require chiptool `device.rs` to emit `uint64` for large base addresses. |
| R8 | QEMU `palmetto-bmc` (AST2400) may not accurately emulate SRAM alias at `0x0` | **Low** — QEMU test may pass but hardware may fail | Use hardware as the primary gate for AST2400 (T3.4). |
| R9 | AST2400/2500 SDRAM remap: boot ROM may remap SDRAM to `0x00000000` before jumping to TamaGo. If so, `goos.RamStart=0x0` (not `0x40000000`/`0x80000000`) and linker `-T=0x00010000`. | **High** — wrong RamStart means heap overlaps exception vectors or stack is wrong | Confirm from `ast2400v14.md`/`ast2500v18.md` §Boot Sequence before setting RamStart in T3.2/T2.5. |
| R10 | ASPEED FMC image format: TamaGo ELF must be packed into the ASPEED FMC boot image format for SPI loading. Format requires specific header (magic, checksum, load/entry addresses). | **High** — without correct format, boot ROM won't load the image | Implement FMC image packager in T1.3 (before hardware test T1.5). Reference: ASPEED SDK or u-boot SPL source for FMC header layout. |
| R11 | `GOARM=6,softfloat` and `GOARM=7,softfloat` do not expose `arm.softfloat` as a Go build tag. The `-tags softfloat` convention used in this plan is user-defined. | **Med** — anyone building without `-tags softfloat` will get VFP enabled unexpectedly on AST2500 (no VFP) causing a fatal exception | Document this clearly in all board package READMEs and the top-level Makefile. |

### Decisions Required Before Implementation

1. **VIC base address** (before T2.3): Read `ast2400v14.md` §4 VIC to determine whether the register bank starts at `0x1e6c0000` or `0x1e6c0080`. This determines the `INTC.Base` constant used across AST2400 and AST2500.

2. **AST2400 exception vector placement** (before T3.2): Confirm from `ast2400v14.md` §2 Memory Map whether address `0x00000000` maps to internal SRAM (allowing exception stubs to be installed there from cpuinit.s) or requires an SCU remap register to be set first.

3. **chiptool `uint64` for large base addresses** (before T4.2): Decide whether to patch `device.rs` to emit `uint64` constants when the base address exceeds `0xFFFFFFFF`, or to handle it in the YAML/template. This affects the generated API for AST2700 only.

---

## File Creation Summary

### New files (tamago repo)

```
arm/
  features_v5.go              (T0.2 — !arm.6 no-op initFeatures)

soc/aspeed/
  uart/uart.go                (T2.1)
  timer/timer.go              (T2.2)
  timer/timer.s               (T2.2)
  intc/intc.go                (T2.3)

  ast2600/
    regs.yaml                 (T1.1 — hand-written)
    scu.gen.go                (T1.1 — generated)
    uart.gen.go               (T1.1 — generated)
    timer.gen.go              (T1.1 — generated)
    trng.gen.go               (T1.1 — generated)
    device.gen.go             (T1.1 — generated)
    ast2600.go                (T1.2)
    init.go                   (T1.2)
    timer.go                  (T1.2)
    rng.go                    (T1.2)
    mem.go                    (T1.2)
    clock.go                  (T1.2)

  ast2500/
    regs.yaml + 5 *.gen.go   (T2.4)
    ast2500.go  init.go  timer.go  rng.go  mem.go  clock.go  earlyinit.s  (T2.5)

  ast2400/
    regs.yaml + 5 *.gen.go   (T3.1)
    ast2400.go  init.go  timer.go  rng.go  mem.go  clock.go  earlyinit.s  (T3.2)

  ast2700/
    regs.yaml + N *.gen.go   (T4.2)
    ast2700.go  init.go  timer.go  rng.go  mem.go  clock.go  earlyinit.s  (T4.3)

board/aspeed/
  ast2600evb/  cpuinit.s  ast2600evb.go  console.go  mem.go  README.md   (T1.3, T1.5)
  ast2500evb/  cpuinit.s  ast2500evb.go  console.go  mem.go  README.md   (T2.6, T2.7)
  ast2400evb/  cpuinit.s  ast2400evb.go  console.go  mem.go  README.md   (T3.3, T3.4)
  ast2700evb/  cpuinit.s  ast2700evb.go  console.go  mem.go  README.md   (T4.4, T4.5)
```

### Modified files (tamago repo)

```
arm/features.go              (T0.2 — add //go:build arm.6)
arm/arm.go                   (T0.2 — revert Init() to no-arg, remove NoVBAR guard)
soc/nuvoton/nuc980/nuc980.go (T0.2 — ARM.Init() call update)
```

### Modified files (chiptool repo)

```
src/generate/go/mod.rs       (T0.1 — filename generation)
src/commands/generate.rs     (T0.1 — --package, --reg-import CLI args)
```

---

## Build Matrix Reference

| SoC | Command |
|---|---|
| AST2400 | `GOOS=tamago GOARCH=arm GOARM=5 go tool tamago build -tags linkcpuinit -ldflags "-T 0x40010000 -R 0x1000" ./cmd/ast2400test/` |
| AST2500 | `GOOS=tamago GOARCH=arm GOARM=6,softfloat go tool tamago build -tags linkcpuinit,softfloat -ldflags "-T 0x80010000 -R 0x1000" ./cmd/ast2500test/` |
| AST2600 | `GOOS=tamago GOARCH=arm GOARM=7,softfloat go tool tamago build -tags linkcpuinit,softfloat -ldflags "-T 0x80010000 -R 0x1000" ./cmd/ast2600test/` |
| AST2700 | `GOOS=tamago GOARCH=arm64 go tool tamago build -tags linkcpuinit -ldflags "-T 0x400010000 -R 0x1000" ./cmd/ast2700test/` |

---
*This plan was generated on 2026-04-27 based on the aspeed branch state, NUC980 PR #66 contents, AST2400/2500/2600/2700 datasheets, and the ASPEED u-boot vendor fork (branch aspeed-master-v2023.10).*
