// Copyright (c) 2023-2025, Nubificus LTD
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hypervisors

/*
#cgo LDFLAGS: -L/usr/local/lib64 -lkrun
#cgo CFLAGS: -I/usr/local/include
#include <libkrun.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/urunc-dev/urunc/pkg/unikontainers/types"
)

const (
	KrunVmm    VmmType = "libkrun"
	KrunBinary string  = "libkrun"
)

type Krun struct {
	binaryPath string
	binary     string
}

// Stop kills the libkrun VM process
func (k *Krun) Stop(pid int) error {
	return killProcess(pid)
}

// UsesKVM returns true as libkrun uses KVM
func (k *Krun) UsesKVM() bool {
	return true
}

// SupportsSharedfs returns false as libkrun has limited device support
func (k *Krun) SupportsSharedfs(_ string) bool {
	return false
}

// Path returns the path to libkrun
func (k *Krun) Path() string {
	return k.binaryPath
}

// Ok checks if libkrun is available
func (k *Krun) Ok() error {
	// Check if libkrun.so is loadable by attempting to create a context
	ctxID := C.krun_create_ctx()
	if ctxID < 0 {
		return ErrVMMNotInstalled
	}
	C.krun_free_ctx(C.uint(ctxID))
	return nil
}

func (k *Krun) Execve(args types.ExecArgs, ukernel types.Unikernel) error {
	vmmLog.Debug("Starting libkrun VM configuration")

	// Create libkrun context
	ctxID := C.krun_create_ctx()
	if ctxID < 0 {
		return fmt.Errorf("krun_create_ctx failed with error code: %d", ctxID)
	}
	defer C.krun_free_ctx(C.uint(ctxID))

	// Set VM config (memory and vCPUs)
	ramMiB := C.uint(args.MemSizeB / (1024 * 1024))
	if ramMiB == 0 {
		ramMiB = C.uint(DefaultMemory)
	}
	numVCPUs := C.uchar(args.VCPUs)
	if numVCPUs == 0 {
		numVCPUs = 1
	}

	ret := C.krun_set_vm_config(C.uint(ctxID), numVCPUs, ramMiB)
	if ret < 0 {
		return fmt.Errorf("krun_set_vm_config failed with error code: %d", ret)
	}
	vmmLog.Debugf("Set VM config: %d vCPUs, %d MiB RAM", numVCPUs, ramMiB)

	// Set root filesystem from sharedfs (if available)
	if args.Sharedfs.Path != "" {
		cRoot := C.CString(args.Sharedfs.Path)
		defer C.free(unsafe.Pointer(cRoot))
		ret = C.krun_set_root(C.uint(ctxID), cRoot)
		if ret < 0 {
			return fmt.Errorf("krun_set_root failed with error code: %d", ret)
		}
		vmmLog.Debugf("Set root: %s", args.Sharedfs.Path)
	}

	// Add kernel if provided
	if args.UnikernelPath != "" {
		cKernel := C.CString(args.UnikernelPath)
		defer C.free(unsafe.Pointer(cKernel))
		
		// Set kernel with optional initrd and command line
		var cInitrd *C.char
		if args.InitrdPath != "" {
			cInitrd = C.CString(args.InitrdPath)
			defer C.free(unsafe.Pointer(cInitrd))
		}

		// Build command line from args.Command (which is a string)
		var cCmdline *C.char
		if args.Command != "" {
			cCmdline = C.CString(args.Command)
			defer C.free(unsafe.Pointer(cCmdline))
		}

		// kernel_format: 0 for default/auto-detect
		ret = C.krun_set_kernel(C.uint(ctxID), cKernel, 0, cInitrd, cCmdline)
		if ret < 0 {
			return fmt.Errorf("krun_set_kernel failed with error code: %d", ret)
		}
		if args.InitrdPath != "" {
			vmmLog.Debugf("Set kernel: %s, initrd: %s", args.UnikernelPath, args.InitrdPath)
		} else {
			vmmLog.Debugf("Set kernel: %s", args.UnikernelPath)
		}
	}

	// Add block devices
	blockArgs := ukernel.MonitorBlockCli()
	for _, blockArg := range blockArgs {
		cBlockID := C.CString(blockArg.ID)
		cBlockPath := C.CString(blockArg.Path)
		defer C.free(unsafe.Pointer(cBlockID))
		defer C.free(unsafe.Pointer(cBlockPath))
		
		ret = C.krun_add_disk(C.uint(ctxID), cBlockID, cBlockPath, C.bool(false))
		if ret < 0 {
			return fmt.Errorf("krun_add_disk failed for %s with error code: %d", blockArg.ID, ret)
		}
		vmmLog.Debugf("Added block device: %s -> %s", blockArg.ID, blockArg.Path)
	}

	// Configure networking if tap device provided
	if args.Net.TapDev != "" {
		cTapDev := C.CString(args.Net.TapDev)
		defer C.free(unsafe.Pointer(cTapDev))
		
		// krun_add_net_tap takes (ctx_id, tap_name, mac, features, flags)
		// Pass nil for mac to use default, 0 for features/flags
		ret = C.krun_add_net_tap(C.uint(ctxID), cTapDev, nil, 0, 0)
		if ret < 0 {
			return fmt.Errorf("krun_add_net_tap failed with error code: %d", ret)
		}
		vmmLog.Debugf("Set network tap device: %s", args.Net.TapDev)

		// Set MAC address if provided
		if args.Net.MAC != "" {
			// Parse MAC address to byte array
			macStr := strings.ReplaceAll(args.Net.MAC, ":", "")
			if len(macStr) == 12 {
				var macBytes [6]C.uint8_t
				for i := 0; i < 6; i++ {
					var b byte
					fmt.Sscanf(macStr[i*2:i*2+2], "%02x", &b)
					macBytes[i] = C.uint8_t(b)
				}
				ret = C.krun_set_net_mac(C.uint(ctxID), &macBytes[0])
				if ret < 0 {
					return fmt.Errorf("krun_set_net_mac failed with error code: %d", ret)
				}
				vmmLog.Debugf("Set network MAC: %s", args.Net.MAC)
			}
		}
	}

	// Set environment variables
	// krun_set_env expects a null-terminated array of C strings
	if len(args.Environment) > 0 {
		// Create null-terminated array of environment variables
		cEnv := make([]*C.char, len(args.Environment)+1)
		for i, env := range args.Environment {
			cEnv[i] = C.CString(env)
			defer C.free(unsafe.Pointer(cEnv[i]))
		}
		cEnv[len(args.Environment)] = nil // null-terminate

		ret = C.krun_set_env(C.uint(ctxID), &cEnv[0])
		if ret < 0 {
			vmmLog.Warnf("krun_set_env failed with error code: %d", ret)
		}
	}

	// Start the VM
	vmmLog.Debug("Starting libkrun VM")
	ret = C.krun_start_enter(C.uint(ctxID))
	if ret < 0 {
		return fmt.Errorf("krun_start_enter failed with error code: %d", ret)
	}

	// krun_start_enter blocks until VM exits
	vmmLog.Debug("libkrun VM exited")
	os.Exit(0)
	return nil
}
