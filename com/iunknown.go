//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var IID_IUnknown = &windows.GUID{
	Data1: 0x00000000,
	Data2: 0x0000,
	Data3: 0x0000,
	Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
}

// IUnknownVtbl is the virtual function table for the IUnknown interface.
// It contains pointers to the three fundamental COM methods: QueryInterface, AddRef, and Release.
type IUnknownVtbl struct {
	// QueryInterface looks up a pointer to another interface on the object.
	QueryInterface uintptr
	// AddRef increments the reference count for an interface on an object.
	AddRef uintptr
	// Release decrements the reference count for an interface on an object.
	Release uintptr
}

// IUnknown is the base interface for all other COM interfaces.
// It provides the fundamental mechanisms for object lifetime management (AddRef/Release)
// and interface discovery (QueryInterface).
type IUnknown struct {
	// LpVtbl is a pointer to the virtual function table for this interface.
	LpVtbl *[1024]uintptr
}

func (v *IUnknown) Vtbl() *IUnknownVtbl {
	return (*IUnknownVtbl)(unsafe.Pointer(v.LpVtbl))
}

// QueryInterface looks up a pointer to another interface on the object.
//
// HRESULT QueryInterface(
//
//	REFIID riid,
//	void   **ppvObject
//
// );
func (v *IUnknown) QueryInterface(riid *windows.GUID, ppvObject unsafe.Pointer) (ret error) {
	r0, _, _ := syscall.SyscallN(v.Vtbl().QueryInterface, uintptr(unsafe.Pointer(v)), uintptr(unsafe.Pointer(riid)), uintptr(ppvObject))
	if r0 != 0 {
		ret = syscall.Errno(r0)
	}
	return
}

// Release decrements the reference count for an interface on an object.
//
// ULONG Release();
func (v *IUnknown) Release() uint32 {
	ret, _, _ := syscall.SyscallN(v.Vtbl().Release, uintptr(unsafe.Pointer(v)))
	return uint32(ret)
}
