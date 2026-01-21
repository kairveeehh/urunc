// Copyright (c) 2023-2025, Nubificus LTD
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hypervisors

/*
#cgo LDFLAGS: -lkrun
#include <libkrun.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/urunc-dev/urunc/pkg/unikontainers/types"
	log "github.com/sirupsen/logrus"
)

const (
	LibkrunVmm VmmType = "libkrun"
)

type Libkrun struct {
	// Libkrun is a library, so we don't need a binary path like QEMU
}

func (l *Libkrun) Stop(pid int) error {
	return killProcess(pid)
}

func (l *Libkrun) Ok() error {
	return nil
}

func (l *Libkrun) UsesKVM() bool {
	return true
}

func (l *Libkrun) SupportsSharedfs(_ string) bool {
	return false 
}

func (l *Libkrun) Path() string {
	return "internal-library"
}

func (l *Libkrun) Execve(args types.ExecArgs, ukernel types.Unikernel) error {
	log.Debug("Initializing libkrun context...")

	// 1. Convert Memory (Bytes -> MB)
	memMB := BytesToStringMB(args.MemSizeB)
	var ramMib uint32
	fmt.Sscanf(memMB, "%d", &ramMib) 

	// 2. Create Context
	rawCtxId := C.krun_create_ctx()
	if rawCtxId < 0 {
		return fmt.Errorf("failed to create libkrun context")
	}
	// FIX 1: Explicitly cast to C.uint32_t for all future calls
	ctxId := C.uint32_t(rawCtxId)

	// 3. Configure VM (vCPUs, RAM)
	res := C.krun_set_vm_config(ctxId, C.uint8_t(args.VCPUs), C.uint32_t(ramMib))
	if res < 0 {
		return fmt.Errorf("failed to set VM config (CPUs: %d, RAM: %dMB)", args.VCPUs, ramMib)
	}

	// FIX 2: Set Kernel/Initrd/Cmdline separately BEFORE starting
	
	// Set Kernel
	cKernel := C.CString(args.UnikernelPath)
	defer C.free(unsafe.Pointer(cKernel))
	if ret := C.krun_set_kernel(ctxId, cKernel); ret < 0 {
		return fmt.Errorf("failed to set kernel path: %s", args.UnikernelPath)
	}

	// Set Initrd (if present)
	if args.InitrdPath != "" {
		cInitrd := C.CString(args.InitrdPath)
		defer C.free(unsafe.Pointer(cInitrd))
		if ret := C.krun_set_initrd(ctxId, cInitrd); ret < 0 {
			return fmt.Errorf("failed to set initrd: %s", args.InitrdPath)
		}
	}

	// Set Command Line arguments
	if args.Command != "" {
		cCmdline := C.CString(args.Command)
		defer C.free(unsafe.Pointer(cCmdline))
		if ret := C.krun_set_cmdline(ctxId, cCmdline); ret < 0 {
			return fmt.Errorf("failed to set cmdline: %s", args.Command)
		}
	}

	// 4. Networking (Placeholder - skipped to prevent errors for now)
	if args.Net.TapDev != "" {
		log.Warn("Networking requires verification of libkrun network API")
	}

	log.WithField("libkrun", "start").Info("Starting VM via library call")

	// 5. Start the VM
	// FIX 3: Pass ONLY the ctxId, as indicated by the error "want (_Ctype_uint32_t)"
	ret := C.krun_start_enter(ctxId)

	log.Infof("Libkrun VM exited with code: %d", ret)
	os.Exit(int(ret))
	
	return nil
}