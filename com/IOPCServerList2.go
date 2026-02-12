//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var IID_IOPCServerList2 = windows.GUID{
	Data1: 0x9DD0B56C,
	Data2: 0xAD9E,
	Data3: 0x43ee,
	Data4: [8]byte{0x83, 0x05, 0x48, 0x7F, 0x31, 0x88, 0xBF, 0x7A},
}

// IOPCServerList2 is an updated version of IOPCServerList providing additional details, as defined in the OPC Data Access Custom Interface Standard.
type IOPCServerList2 struct {
	// IUnknown is the underlying COM interface.
	*IUnknown
}

func (sl *IOPCServerList2) Vtbl() *IOPCServerListVtbl {
	return (*IOPCServerListVtbl)(unsafe.Pointer(sl.IUnknown.LpVtbl))
}

// EnumClassesOfCategories enumerates OPC servers belonging to specified categories.
//
// Example:
//
//	enum, err := list2.EnumClassesOfCategories([]windows.GUID{com.OPCCAT_DA20}, nil)
func (sl *IOPCServerList2) EnumClassesOfCategories(rgcatidImpl []windows.GUID, rgcatidReq []windows.GUID) (ppenumClsid *IEnumGUID, err error) {
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

// GetClassDetails retrieves ProgID, UserType, and Version Independent ProgID for a given CLSID.
//
// Example:
//
//	pProgID, pUserType, pVIPID, err := list2.GetClassDetails(&clsid)
func (sl *IOPCServerList2) GetClassDetails(guid *windows.GUID) (*uint16, *uint16, *uint16, error) {
	var ppszProgID, ppszUserType, ppszVerIndProgID *uint16
	r0, _, _ := syscall.SyscallN(sl.Vtbl().GetClassDetails, uintptr(unsafe.Pointer(sl.IUnknown)), uintptr(unsafe.Pointer(guid)), uintptr(unsafe.Pointer(&ppszProgID)), uintptr(unsafe.Pointer(&ppszUserType)), uintptr(unsafe.Pointer(&ppszVerIndProgID)))
	if r0 != 0 {
		return nil, nil, nil, syscall.Errno(r0)
	}
	return ppszProgID, ppszUserType, ppszVerIndProgID, nil
}

// CLSIDFromProgID retrieves the CLSID for a given ProgID.
//
// Example:
//
//	clsid, err := list2.CLSIDFromProgID("Matrikon.OPC.Simulation.1")
func (sl *IOPCServerList2) CLSIDFromProgID(szProgID string) (*windows.GUID, error) {
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
