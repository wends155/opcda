//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// IEnumGUIDVtbl is the virtual function table for the IEnumGUID interface.
// It inherits from IUnknownVtbl and adds methods for enumerating GUIDs.
type IEnumGUIDVtbl struct {
	IUnknownVtbl
	// Next retrieves the next celt items in the enumeration sequence.
	Next uintptr
	// Skip skips over the next celt items in the enumeration sequence.
	Skip uintptr
	// Reset resets the enumeration sequence to the beginning.
	Reset uintptr
	// Clone creates a new enumerator that contains the same enumeration state as the current one.
	Clone uintptr
}

// IEnumGUID is used to enumerate through an array of GUIDs.
type IEnumGUID struct {
	// IUnknown is the underlying COM interface.
	*IUnknown
}

func (ie *IEnumGUID) Vtbl() *IEnumGUIDVtbl {
	return (*IEnumGUIDVtbl)(unsafe.Pointer(ie.IUnknown.LpVtbl))
}

func (ie *IEnumGUID) Next(celt uint32, rgelt *windows.GUID, pceltFetched *uint32) error {
	r0, _, _ := syscall.SyscallN(ie.Vtbl().Next, uintptr(unsafe.Pointer(ie.IUnknown)), uintptr(celt), uintptr(unsafe.Pointer(rgelt)), uintptr(unsafe.Pointer(pceltFetched)))
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}
