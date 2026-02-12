//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var CLSID_OpcServerList = windows.GUID{
	Data1: 0x13486D51,
	Data2: 0x4821,
	Data3: 0x11D2,
	Data4: [8]byte{0xA4, 0x94, 0x3C, 0xB3, 0x06, 0xC1, 0x00, 0x00},
}

var IID_IOPCServerList = windows.GUID{
	Data1: 0x13486D50,
	Data2: 0x4821,
	Data3: 0x11D2,
	Data4: [8]byte{0xA4, 0x94, 0x3C, 0xB3, 0x06, 0xC1, 0x00, 0x00},
}

// IOPCServerListVtbl is the virtual function table for the IOPCServerList interface.
type IOPCServerListVtbl struct {
	IUnknownVtbl
	// EnumClassesOfCategories enumerates OPC servers belonging to specified categories.
	EnumClassesOfCategories uintptr
	// GetClassDetails retrieves ProgID and UserType for a given CLSID.
	GetClassDetails uintptr
	// CLSIDFromProgID retrieves the CLSID for a given ProgID.
	CLSIDFromProgID uintptr
}

// IOPCServerList provides methods to enumerate and find OPC servers as defined in the OPC Data Access Custom Interface Standard.
type IOPCServerList struct {
	// IUnknown is the underlying COM interface.
	*IUnknown
}

func (sl *IOPCServerList) Vtbl() *IOPCServerListVtbl {
	return (*IOPCServerListVtbl)(unsafe.Pointer(sl.IUnknown.LpVtbl))
}

// EnumClassesOfCategories enumerates OPC servers belonging to specified categories.
//
// Example:
//
//	enum, err := list.EnumClassesOfCategories([]windows.GUID{com.OPCCAT_DA20}, nil)
func (sl *IOPCServerList) EnumClassesOfCategories(rgcatidImpl []windows.GUID, rgcatidReq []windows.GUID) (ppenumClsid *IEnumGUID, err error) {
	var r0 uintptr
	cImplemented := uint32(len(rgcatidImpl))
	cRequired := uint32(len(rgcatidReq))
	var iUnknown *IUnknown
	if cRequired == 0 {
		r0, _, _ = syscall.SyscallN(sl.Vtbl().EnumClassesOfCategories, uintptr(unsafe.Pointer(sl.IUnknown)), uintptr(cImplemented), uintptr(unsafe.Pointer(&rgcatidImpl[0])), uintptr(0), uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(&iUnknown)))
	} else {
		r0, _, _ = syscall.SyscallN(sl.Vtbl().EnumClassesOfCategories, uintptr(unsafe.Pointer(sl.IUnknown)), uintptr(cImplemented), uintptr(unsafe.Pointer(&rgcatidImpl[0])), uintptr(cRequired), uintptr(unsafe.Pointer(&rgcatidReq[0])), uintptr(unsafe.Pointer(&iUnknown)))
	}
	if r0 != 0 {
		err = syscall.Errno(r0)
		return
	}
	ppenumClsid = &IEnumGUID{IUnknown: iUnknown}
	return
}

// GetClassDetails retrieves ProgID and UserType for a given CLSID.
//
// Example:
//
//	pProgID, pUserType, err := list.GetClassDetails(&clsid)
func (sl *IOPCServerList) GetClassDetails(guid *windows.GUID) (*uint16, *uint16, error) {
	var ppszProgID, ppszUserType *uint16
	r0, _, _ := syscall.SyscallN(sl.Vtbl().GetClassDetails, uintptr(unsafe.Pointer(sl.IUnknown)), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(&ppszProgID)), uintptr(unsafe.Pointer(&ppszUserType)))
	if r0 != 0 {
		return nil, nil, syscall.Errno(r0)
	}
	return ppszProgID, ppszUserType, nil
}

// CLSIDFromProgID retrieves the CLSID for a given ProgID.
//
// Example:
//
//	clsid, err := list.CLSIDFromProgID("Matrikon.OPC.Simulation.1")
func (sl *IOPCServerList) CLSIDFromProgID(szProgID string) (*windows.GUID, error) {
	var clsid windows.GUID
	pProgID, err := syscall.UTF16PtrFromString(szProgID)
	if err != nil {
		return nil, err
	}
	r0, _, _ := syscall.SyscallN(sl.Vtbl().CLSIDFromProgID, uintptr(unsafe.Pointer(sl.IUnknown)), uintptr(unsafe.Pointer(pProgID)), uintptr(unsafe.Pointer(&clsid)))
	if r0 != 0 {
		return nil, syscall.Errno(r0)
	}
	return &clsid, nil
}
