# ASPEED AST2700 QEMU — TamaGo build and run
#
# Targets:
#   make              — build QEMU ELF
#   make dcscm        — build AST2750-A1 DCSCM ELF
#   make qemu         — QEMU via -kernel (fast, ELF loaded directly)
#   make clean        — remove build artefacts
#
# Required environment / overrides:
#   TAMAGO   path to tamago-go binary   (default: tries well-known dev path then PATH)
#   GOROOT   tamago-go root             (inferred from TAMAGO if not set)
#   QEMU     qemu-system-aarch64 binary (default: qemu-system-aarch64)

QEMU  ?= qemu-system-aarch64

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

TARGET    := ast2700test
ELF       := $(TARGET).elf
DCSCM_ELF := $(TARGET)_dcscm.elf

# AST2700 SDRAM at 0x400000000, text starts at +0x10000
LDFLAGS      := -T 0x400010000 -R 0x1000
DCSCM_LDFLAGS := -T 0x430000000 -R 0x1000
TAGS         := linkcpuinit
DCSCM_TAGS   := linkcpuinit,ast2700dcscm

export GOTOOLCHAIN := local
export GOROOT
export GOOSPKG     := github.com/usbarmory/tamago
export GOOS        := tamago
export GOARCH      := arm64

.PHONY: all dcscm qemu clean

all: $(ELF)

$(ELF):
	PATH=$(GOROOT)/bin:$$PATH \
	$(TAMAGO) build \
	  -tags $(TAGS) \
	  -ldflags "$(LDFLAGS)" \
	  -o $@ \
	  $(SCRIPT_DIR)

dcscm:
	PATH=$(GOROOT)/bin:$$PATH \
	$(TAMAGO) build \
	  -tags $(DCSCM_TAGS) \
	  -ldflags "$(DCSCM_LDFLAGS)" \
	  -o $(DCSCM_ELF) \
	  $(SCRIPT_DIR)

qemu: $(ELF)
	$(QEMU) -M ast2700a2-evb -nographic -kernel $(ELF)

clean:
	rm -f $(ELF) $(DCSCM_ELF)
