//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var IID_IOPCAsyncIO2 = windows.GUID{
	Data1: 0x39c13a71,
	Data2: 0x011e,
	Data3: 0x11d0,
	Data4: [8]byte{0x96, 0x75, 0x00, 0x20, 0xaf, 0xd8, 0xad, 0xb3},
}

type IOPCAsyncIO2Vtbl struct {
	IUnknownVtbl
	Read      uintptr
	Write     uintptr
	Refresh2  uintptr
	Cancel2   uintptr
	SetEnable uintptr
	GetEnable uintptr
}

// IOPCAsyncIO2 provides asynchronous data access to OPC items.
// It uses connection points for call-backs to the client.
type IOPCAsyncIO2 struct {
	*IUnknown
}

func (sl *IOPCAsyncIO2) Vtbl() *IOPCAsyncIO2Vtbl {
	return (*IOPCAsyncIO2Vtbl)(unsafe.Pointer(sl.IUnknown.LpVtbl))
}

// Read performs an asynchronous read of one or more items in the group.
// The results are returned via the IOPCDataCallback interface.
//
// Parameters:
//
//	phServer: Server handles of the items to read.
//	dwTransactionID: A client-generated transaction ID.
//
// Returns:
//
//	pdwCancelID: A cancel ID that can be used to cancel the read.
//	ppErrors: A slice of HRESULTs for each item.
//
// Example:
//
//	cancelID, errors, err := asyncIO.Read(serverHandles, 123)
func (sl *IOPCAsyncIO2) Read(phServer []uint32, dwTransactionID uint32) (pdwCancelID uint32, ppErrors []int32, err error) {
	var pErrors unsafe.Pointer
	r0, _, _ := syscall.SyscallN(
		sl.Vtbl().Read,
		uintptr(unsafe.Pointer(sl.IUnknown)),
		uintptr(len(phServer)),
		uintptr(unsafe.Pointer(&phServer[0])),
		uintptr(dwTransactionID),
		uintptr(unsafe.Pointer(&pdwCancelID)),
		uintptr(unsafe.Pointer(&pErrors)))
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	defer func() {
		if pErrors != nil {
			CoTaskMemFree(pErrors)
		}
	}()
	ppErrors = make([]int32, len(phServer))
	for i := uint32(0); i < uint32(len(phServer)); i++ {
		errNo := *(*int32)(unsafe.Pointer(uintptr(pErrors) + uintptr(i)*4))
		ppErrors[i] = int32(errNo)
	}
	return
}

// Write performs an asynchronous write of one or more items in the group.
//
// Example:
//
//	cancelID, errors, err := asyncIO.Write(serverHandles, variants, 456)
func (sl *IOPCAsyncIO2) Write(phServer []uint32, pItemValues []VARIANT, dwTransactionID uint32) (pdwCancelID uint32, ppErrors []int32, err error) {
	var pErrors unsafe.Pointer
	r0, _, _ := syscall.SyscallN(
		sl.Vtbl().Write,
		uintptr(unsafe.Pointer(sl.IUnknown)),
		uintptr(len(phServer)),
		uintptr(unsafe.Pointer(&phServer[0])),
		uintptr(unsafe.Pointer(&pItemValues[0])),
		uintptr(dwTransactionID),
		uintptr(unsafe.Pointer(&pdwCancelID)),
		uintptr(unsafe.Pointer(&pErrors)))
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	defer func() {
		if pErrors != nil {
			CoTaskMemFree(pErrors)
		}
	}()
	ppErrors = make([]int32, len(phServer))
	for i := uint32(0); i < uint32(len(phServer)); i++ {
		errNo := *(*int32)(unsafe.Pointer(uintptr(pErrors) + uintptr(i)*4))
		ppErrors[i] = int32(errNo)
	}
	return
}

// Refresh2 triggers a refresh of all active items in the group.
//
// Example:
//
//	cancelID, err := asyncIO.Refresh2(com.OPC_DS_DEVICE, 789)
func (sl *IOPCAsyncIO2) Refresh2(dwSource OPCDATASOURCE, dwTransactionID uint32) (pdwCancelID uint32, err error) {
	r0, _, _ := syscall.SyscallN(
		sl.Vtbl().Refresh2,
		uintptr(unsafe.Pointer(sl.IUnknown)),
		uintptr(dwSource),
		uintptr(dwTransactionID),
		uintptr(unsafe.Pointer(&pdwCancelID)))
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	return
}

// Cancel2 attempts to cancel an ongoing asynchronous transaction.
//
// Example:
//
//	err := asyncIO.Cancel2(cancelID)
func (sl *IOPCAsyncIO2) Cancel2(dwCancelID uint32) (err error) {
	r0, _, _ := syscall.SyscallN(
		sl.Vtbl().Cancel2,
		uintptr(unsafe.Pointer(sl.IUnknown)),
		uintptr(dwCancelID),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	return
}
