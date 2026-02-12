//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var IID_IOPCItemProperties = windows.GUID{
	Data1: 0x39c13a72,
	Data2: 0x011e,
	Data3: 0x11d0,
	Data4: [8]byte{0x96, 0x75, 0x00, 0x20, 0xaf, 0xd8, 0xad, 0xb3},
}

// IOPCItemProperties allows clients to query the properties of an OPC item as defined in the OPC Data Access Custom Interface Standard.
type IOPCItemProperties struct {
	// IUnknown is the underlying COM interface.
	*IUnknown
}

// IOPCItemPropertiesVtbl is the virtual function table for the IOPCItemProperties interface.
type IOPCItemPropertiesVtbl struct {
	IUnknownVtbl
	// QueryAvailableProperties returns a list of properties available for the specified item.
	QueryAvailableProperties uintptr
	// GetItemProperties retrieves the current values for one or more properties of an item.
	GetItemProperties uintptr
	// LookupItemIDs provides the ItemIDs for one or more properties of an item.
	LookupItemIDs uintptr
}

func (v *IOPCItemProperties) Vtbl() *IOPCItemPropertiesVtbl {
	return (*IOPCItemPropertiesVtbl)(unsafe.Pointer(v.IUnknown.LpVtbl))
}

var pointerSize uintptr = unsafe.Sizeof(uintptr(0))

// QueryAvailableProperties returns a list of properties available for the specified item.
//
// Parameters:
//
//	szItemID: The item ID to query.
//
// Returns:
//
//	A list of property IDs, their descriptions, and their data types.
//
// Example:
//
//	ids, descs, types, err := prop.QueryAvailableProperties("Random.Int4")
func (v *IOPCItemProperties) QueryAvailableProperties(szItemID string) (ppPropertyIDs []uint32, ppDescriptions []string, ppvtDataTypes []uint16, err error) {
	var pPropertyIDs unsafe.Pointer
	var pDescriptions unsafe.Pointer
	var pvtDataTypes unsafe.Pointer
	var pName *uint16
	pName, err = syscall.UTF16PtrFromString(szItemID)
	if err != nil {
		return
	}
	var pdwCount uint32
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().QueryAvailableProperties,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(unsafe.Pointer(pName)),
		uintptr(unsafe.Pointer(&pdwCount)),
		uintptr(unsafe.Pointer(&pPropertyIDs)),
		uintptr(unsafe.Pointer(&pDescriptions)),
		uintptr(unsafe.Pointer(&pvtDataTypes)))
	if r0 != 0 {
		err = syscall.Errno(r0)
		return
	}
	defer func() {
		CoTaskMemFree(pPropertyIDs)
		CoTaskMemFree(pDescriptions)
		CoTaskMemFree(pvtDataTypes)
	}()
	if pdwCount == 0 {
		return
	}
	ppPropertyIDs = make([]uint32, pdwCount)
	ppDescriptions = make([]string, pdwCount)
	ppvtDataTypes = make([]uint16, pdwCount)
	for i := uint32(0); i < pdwCount; i++ {
		ppPropertyIDs[i] = *(*uint32)(unsafe.Pointer(uintptr(pPropertyIDs) + uintptr(i)*4))
		pwstr := *(**uint16)(unsafe.Pointer(uintptr(pDescriptions) + uintptr(i)*pointerSize))
		ppDescriptions[i] = windows.UTF16PtrToString(pwstr)
		CoTaskMemFree(unsafe.Pointer(pwstr))
		ppvtDataTypes[i] = *(*uint16)(unsafe.Pointer(uintptr(pvtDataTypes) + uintptr(i)*2))
	}
	return
}

// GetItemProperties retrieves the current values for one or more properties of an item.
//
// Parameters:
//
//	szItemID: The item ID to query.
//	propertyIDs: The IDs of the properties to retrieve.
//
// Example:
//
//	values, errors, err := prop.GetItemProperties("Random.Int4", []uint32{1, 2})
func (v *IOPCItemProperties) GetItemProperties(szItemID string, propertyIDs []uint32) (ppvData []interface{}, ppErrors []int32, err error) {
	var pData unsafe.Pointer
	var pErrors unsafe.Pointer
	var pName *uint16
	pName, err = syscall.UTF16PtrFromString(szItemID)
	if err != nil {
		return
	}
	count := len(propertyIDs)
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().GetItemProperties,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(unsafe.Pointer(pName)),
		uintptr(count),
		uintptr(unsafe.Pointer(&propertyIDs[0])),
		uintptr(unsafe.Pointer(&pData)),
		uintptr(unsafe.Pointer(&pErrors)))
	if r0 != 0 {
		return nil, nil, syscall.Errno(r0)
	}
	defer func() {
		CoTaskMemFree(pData)
		CoTaskMemFree(pErrors)
	}()
	if count == 0 {
		return
	}
	ppvData = make([]interface{}, count)
	ppErrors = make([]int32, count)
	for i := 0; i < count; i++ {
		variant := *(*VARIANT)(unsafe.Pointer(uintptr(pData) + uintptr(i)*unsafe.Sizeof(VARIANT{})))
		errNo := *(*int32)(unsafe.Pointer(uintptr(pErrors) + uintptr(i)*4))
		if errNo >= 0 {
			v, err := variant.Value()
			if err != nil {
				errNo = int32(0x80004005 - 0x100000000) // E_FAIL
				ppvData[i] = nil
			} else {
				ppvData[i] = v
			}
		}
		variant.Clear()
		ppErrors[i] = int32(errNo)
	}
	return
}

// LookupItemIDs provides the ItemIDs for one or more properties of an item.
//
// Example:
//
//	itemIDs, errors, err := prop.LookupItemIDs("Random.Int4", []uint32{1})
func (v *IOPCItemProperties) LookupItemIDs(szItemID string, propertyIDs []uint32) (ppszNewItemIDs []string, ppErrors []int32, err error) {
	var pNewIDs unsafe.Pointer
	var pErrors unsafe.Pointer
	var pName *uint16
	pName, err = syscall.UTF16PtrFromString(szItemID)
	if err != nil {
		return
	}
	count := len(propertyIDs)
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().LookupItemIDs,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(unsafe.Pointer(pName)),
		uintptr(count),
		uintptr(unsafe.Pointer(&propertyIDs[0])),
		uintptr(unsafe.Pointer(&pNewIDs)),
		uintptr(unsafe.Pointer(&pErrors)))
	if int32(r0) < 0 {
		return nil, nil, syscall.Errno(r0)
	}
	defer func() {
		CoTaskMemFree(pNewIDs)
		CoTaskMemFree(pErrors)
	}()
	if count == 0 {
		return
	}
	ppszNewItemIDs = make([]string, count)
	ppErrors = make([]int32, count)
	for i := 0; i < count; i++ {
		errNo := *(*int32)(unsafe.Pointer(uintptr(pErrors) + uintptr(i)*4))
		ppErrors[i] = int32(errNo)
		if errNo < 0 {
			continue
		}
		pwstr := *(**uint16)(unsafe.Pointer(uintptr(pNewIDs) + uintptr(i)*pointerSize))
		ppszNewItemIDs[i] = windows.UTF16PtrToString(pwstr)
		CoTaskMemFree(unsafe.Pointer(pwstr))
	}
	return
}
