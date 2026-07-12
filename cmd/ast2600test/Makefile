# ASPEED AST2600EVB — TamaGo build and flash image creation
#
# Targets:
#   make              — build ELF
#   make bin          — build raw binary (note sections stripped)
#   make flash        — build 64 MB MTD flash image
#   make qemu         — QEMU via -kernel (fast, ELF loaded directly)
#   make qemu-flash   — QEMU booting from MTD flash image (matches hardware flow)
#   make clean        — remove build artefacts
#
# Required environment / overrides:
#   TAMAGO   path to tamago-go binary   (default: tries well-known dev path then PATH)
#   GOROOT   tamago-go root             (inferred from TAMAGO if not set)
#   CROSS    cross-toolchain prefix     (default: arm-none-eabi-)
#   QEMU     qemu-system-arm binary     (default: qemu-system-arm)

CROSS ?= arm-none-eabi-
QEMU  ?= qemu-system-arm

# Locate the tamago-go toolchain.
ifeq ($(TAMAGO),)
  _TAMAGO_CANDIDATE := $(HOME)/private/tamago/tamago-go/bin/go
  ifneq ($(wildcard $(_TAMAGO_CANDIDATE)),)
    TAMAGO := $(_TAMAGO_CANDIDATE)
  else
    TAMAGO := go
  endif
endif

ifeq ($(GOROOT),)
  GOROOT := $(abspath $(dir $(TAMAGO))/..)
endif

SCRIPT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

TARGET   := ast2600test
ELF      := $(TARGET).elf
BIN      := $(TARGET).bin
IMG      := $(TARGET).img

LDFLAGS    := -T 0x80010000 -R 0x1000
TAGS       := linkcpuinit,softfloat
FLASH_SIZE := 0x4000000   # 64 MB (w25q512jv, QEMU ast2600-evb default)

export GOTOOLCHAIN := local
export GOROOT
export GOOSPKG     := github.com/usbarmory/tamago
export GOOS        := tamago
export GOARCH      := arm
export GOARM       := 7,softfloat

# ELF entry point — extracted in recipes (ELF does not exist at parse time).
_entry = $(shell $(CROSS)readelf -h $(1) | awk '/Entry point/{print $$4}')

.PHONY: all bin flash qemu qemu-flash clean

all: $(ELF)

$(ELF):
	PATH=$(GOROOT)/bin:$$PATH \
	$(TAMAGO) build \
	  -tags $(TAGS) \
	  -ldflags "$(LDFLAGS)" \
	  -o $@ \
	  $(SCRIPT_DIR)

# Strip Go/GNU note sections so the binary starts cleanly at the .text VMA
# (0x80010000).  The note sections (.note.go.pvh, .note.go.buildid,
# .note.gnu.build-id) are metadata only — the runtime does not read them.
bin: $(BIN)

$(BIN): $(ELF)
	$(CROSS)objcopy -O binary \
	  --remove-section=.note.go.pvh \
	  --remove-section=.note.go.buildid \
	  --remove-section=.note.gnu.build-id \
	  $< $@

flash: $(IMG)

$(IMG): $(BIN) $(ELF)
	python3 $(SCRIPT_DIR)mkflash.py $< $@ $(FLASH_SIZE) $(call _entry,$(ELF))

# QEMU: fast path — ELF loaded directly (QEMU sets load addresses and entry
# point from the ELF headers).  Use for rapid development iterations.
qemu: $(ELF)
	$(QEMU) -M ast2600-evb -nographic -kernel $(ELF)

# QEMU: flash path — mimics the hardware boot flow.
#
#   The MTD flash image is attached at FMC CS0 (0x20000000).  The QEMU
#   internal ROM at 0x00000000 branches to 0x20000000 where the loader stub
#   runs.  However, QEMU's FMC/SPI emulation is too slow for a 1.5 MB
#   bulk copy from memory-mapped flash.
#
#   To work around this, the Makefile pre-loads the binary and sets both
#   CPU reset vectors explicitly:
#
#     -device loader,force-raw=on,addr=0x80010000,file=$(BIN)
#         Writes the raw binary into guest DRAM at 0x80010000 before the
#         machine starts.
#
#     -device loader,cpu-num=0,addr=$(ENTRY)
#     -device loader,cpu-num=1,addr=$(ENTRY)
#         Sets both Cortex-A7 reset vectors to the ELF entry point.
#         CPU1 hits the MPIDR check in cpuinit.s and parks itself.
#
#   Because the CPUs start directly at the entry point the stub never
#   executes; the flash image is present only as it would be on hardware.
#   On real hardware there is no -device loader; the stub runs the full
#   copy path via hardware SPI (~30 ms for 1.5 MB at 100 MHz).
qemu-flash: $(IMG) $(BIN)
	$(QEMU) -M ast2600-evb -nographic \
	  -drive file=$(IMG),format=raw,if=mtd \
	  -device loader,force-raw=on,addr=0x80010000,file=$(BIN) \
	  -device loader,cpu-num=0,addr=$(call _entry,$(ELF)) \
	  -device loader,cpu-num=1,addr=$(call _entry,$(ELF))

clean:
	rm -f $(ELF) $(BIN) $(IMG)
