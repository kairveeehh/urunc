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

// Add the constant for the new driver
const (
	LibkrunVmm VmmType = "libkrun"
)

type Libkrun struct {
	// Libkrun is a library, so we don't need a binary path like QEMU
}

func (l *Libkrun) Stop(pid int) error {
	// In libkrun, the VM is the process. Killing the process stops the VM.
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
	ctxId := C.krun_create_ctx()
	if ctxId < 0 {
		return fmt.Errorf("failed to create libkrun context")
	}

	// 3. Configure VM (vCPUs, RAM)
	res := C.krun_set_vm_config(ctxId, C.uint8_t(args.VCPUs), C.uint32_t(ramMib))
	if res < 0 {
		return fmt.Errorf("failed to set VM config (CPUs: %d, RAM: %dMB)", args.VCPUs, ramMib)
	}

	// 4. Networking (Placeholder)
	if args.Net.TapDev != "" {
		log.Warn("Networking in libkrun requires verifying specific API version methods")
	}

	// 5. Prepare Strings for Start
	cKernel := C.CString(args.UnikernelPath)
	defer C.free(unsafe.Pointer(cKernel))

	var cInitrd *C.char
	if args.InitrdPath != "" {
		cInitrd = C.CString(args.InitrdPath)
		defer C.free(unsafe.Pointer(cInitrd))
	}

	cCmdline := C.CString(args.Command)
	defer C.free(unsafe.Pointer(cCmdline))

	log.WithField("libkrun", "start").Info("Starting VM via library call")

	// 6. Start the VM (This blocks until VM exit)
	ret := C.krun_start_enter(ctxId, cKernel, cInitrd, cCmdline)

	log.Infof("Libkrun VM exited with code: %d", ret)
	
	// Exit the process with the VM's exit code
	os.Exit(int(ret))
	
	return nil
}