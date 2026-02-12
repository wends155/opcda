//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type IEnumStringVtbl struct {
	IUnknownVtbl
	Next  uintptr
	Skip  uintptr
	Reset uintptr
	Clone uintptr
}

// IEnumString is a standard COM interface for enumerating strings.
type IEnumString struct {
	*IUnknown
}

func (sl *IEnumString) Vtbl() *IEnumStringVtbl {
	return (*IEnumStringVtbl)(unsafe.Pointer(sl.IUnknown.LpVtbl))
}

// Next retrieves the next celt items in the enumeration sequence.
//
// Parameters:
//
//	celt: The number of items to retrieve.
//
// Returns:
//
//	A slice of strings containing the retrieved items.
//
// Example:
//
//	items, err := enum.Next(10)
func (sl *IEnumString) Next(celt uint32) (result []string, err error) {
	pRgelt := make([]*uint16, celt)
	var pceltFetched uint32
	r0, _, _ := syscall.SyscallN(
		sl.Vtbl().Next,
		uintptr(unsafe.Pointer(sl.IUnknown)),
		uintptr(celt),
		uintptr(unsafe.Pointer(&pRgelt[0])),
		uintptr(unsafe.Pointer(&pceltFetched)),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	result = make([]string, pceltFetched)
	for i := uint32(0); i < pceltFetched; i++ {
		pwstr := pRgelt[i]
		result[i] = windows.UTF16PtrToString(pwstr)
		CoTaskMemFree(unsafe.Pointer(pwstr))
	}
	return
}
