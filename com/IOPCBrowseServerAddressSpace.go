//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var IID_IOPCBrowseServerAddressSpace = windows.GUID{
	Data1: 0x39c13a4f,
	Data2: 0x011e,
	Data3: 0x11d0,
	Data4: [8]byte{0x96, 0x75, 0x00, 0x20, 0xaf, 0xd8, 0xad, 0xb3},
}

// IOPCBrowseServerAddressSpace provides methods to browse the server's address space.
type IOPCBrowseServerAddressSpace struct {
	*IUnknown
}

type IOPCBrowseServerAddressSpaceVtbl struct {
	IUnknownVtbl
	QueryOrganization    uintptr
	ChangeBrowsePosition uintptr
	BrowseOPCItemIDs     uintptr
	GetItemID            uintptr
	BrowseAccessPaths    uintptr
}

func (v *IOPCBrowseServerAddressSpace) Vtbl() *IOPCBrowseServerAddressSpaceVtbl {
	return (*IOPCBrowseServerAddressSpaceVtbl)(unsafe.Pointer(v.IUnknown.LpVtbl))
}

type OPCNAMESPACETYPE uint32

// QueryOrganization retrieves the organization of the server's address space (hierarchical or flat).
//
// Example:
//
//	org, err := browse.QueryOrganization()
func (v *IOPCBrowseServerAddressSpace) QueryOrganization() (pNameSpaceType OPCNAMESPACETYPE, err error) {
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().QueryOrganization,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(unsafe.Pointer(&pNameSpaceType)),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
	}
	return
}

type OPCBROWSEDIRECTION uint32

// ChangeBrowsePosition changes the current browse position in the address space.
//
// Example:
//
//	err := browse.ChangeBrowsePosition(com.OPC_BROWSE_DOWN, "Folder1")
func (v *IOPCBrowseServerAddressSpace) ChangeBrowsePosition(dwBrowseDirection OPCBROWSEDIRECTION, szString string) (err error) {
	var pName *uint16
	pName, err = syscall.UTF16PtrFromString(szString)
	if err != nil {
		return
	}

	r0, _, _ := syscall.SyscallN(
		v.Vtbl().ChangeBrowsePosition,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(dwBrowseDirection),
		uintptr(unsafe.Pointer(pName)),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
	}
	return
}

type OPCBROWSETYPE uint32

// BrowseOPCItemIDs returns a list of item IDs based on the current browse position and filters.
//
// Parameters:
//
//	dwBrowseFilterType: The type of items to browse (branches, leaves, or both).
//	szFilterCriteria: A filter string (e.g., "*").
//	vtDataTypeFilter: An optional data type filter.
//	dwAccessRightsFilter: An optional access rights filter.
//
// Example:
//
//	items, err := browse.BrowseOPCItemIDs(com.OPC_LEAF, "*", 0, 0)
func (v *IOPCBrowseServerAddressSpace) BrowseOPCItemIDs(dwBrowseFilterType OPCBROWSETYPE, szFilterCriteria string, vtDataTypeFilter uint16, dwAccessRightsFilter uint32) (result []string, err error) {
	var pString *IUnknown
	var pName *uint16
	pName, err = syscall.UTF16PtrFromString(szFilterCriteria)
	if err != nil {
		return
	}

	r0, _, _ := syscall.SyscallN(
		v.Vtbl().BrowseOPCItemIDs,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(dwBrowseFilterType),
		uintptr(unsafe.Pointer(pName)),
		uintptr(vtDataTypeFilter),
		uintptr(dwAccessRightsFilter),
		uintptr(unsafe.Pointer(&pString)),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	ppIEnumString := &IEnumString{pString}
	defer func() {
		ppIEnumString.Release()
	}()

	for {
		var batch []string
		batch, err = ppIEnumString.Next(100)
		if err != nil {
			return nil, err
		}
		if len(batch) < 100 {
			result = append(result, batch...)
			break
		}
		result = append(result, batch...)
	}
	return result, nil
}

// GetItemID retrieves the full item ID for a given browser item name.
//
// Example:
//
//	itemID, err := browse.GetItemID("Item1")
func (v *IOPCBrowseServerAddressSpace) GetItemID(szItemDataID string) (szItemID string, err error) {
	var pString *uint16
	var pName *uint16
	pName, err = syscall.UTF16PtrFromString(szItemDataID)
	if err != nil {
		return
	}
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().GetItemID,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(unsafe.Pointer(pName)),
		uintptr(unsafe.Pointer(&pString)),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	defer func() {
		if pString != nil {
			CoTaskMemFree(unsafe.Pointer(pString))
		}
	}()
	szItemID = windows.UTF16PtrToString(pString)
	return
}
