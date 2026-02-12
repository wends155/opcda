// Package com provides thin wrappers for Windows COM and OLE Automation.
// It is specifically tailored for interacting with OPC DA (Data Access) servers.
//go:build windows

package com

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modOle32                    = windows.NewLazySystemDLL("ole32.dll")
	procCoCreateInstanceEx      = modOle32.NewProc("CoCreateInstanceEx")
	procCoInitializeSecurity    = modOle32.NewProc("CoInitializeSecurity")
	modOleaut32                 = windows.NewLazySystemDLL("oleaut32.dll")
	procVariantClear            = modOleaut32.NewProc("VariantClear")
	procVariantTimeToSystemTime = modOleaut32.NewProc("VariantTimeToSystemTime")
	procSystemTimeToVariantTime = modOleaut32.NewProc("SystemTimeToVariantTime")
	procSafeArrayGetVarType     = modOleaut32.NewProc("SafeArrayGetVartype")
	procSafeArrayGetLBound      = modOleaut32.NewProc("SafeArrayGetLBound")
	procSafeArrayGetUBound      = modOleaut32.NewProc("SafeArrayGetUBound")
	procSafeArrayGetElement     = modOleaut32.NewProc("SafeArrayGetElement")
	procSysAllocStringLen       = modOleaut32.NewProc("SysAllocStringLen")
	procSafeArrayCreateVector   = modOleaut32.NewProc("SafeArrayCreateVector")
	procSafeArrayPutElement     = modOleaut32.NewProc("SafeArrayPutElement")
	procSysFreeString           = modOleaut32.NewProc("SysFreeString")
)

// CoTaskMemFree frees a block of memory previously allocated through a call to CoTaskMemAlloc or CoTaskMemRealloc.
// It is essential for freeing memory returned by COM methods that allocate memory for the caller.
//
// Example:
//
//	var p unsafe.Pointer = ... // from a COM call
//	defer com.CoTaskMemFree(p)
func CoTaskMemFree(pv unsafe.Pointer) {
	if pv == nil {
		return
	}
	windows.CoTaskMemFree(pv)
}

type CLSCTX uint32

const (
	CLSCTX_LOCAL_SERVER  CLSCTX = 0x4
	CLSCTX_REMOTE_SERVER CLSCTX = 0x10
)

// COAUTHIDENTITY contains a user name and password.
// It is used to establish a non-default client identity for authentication.
type COAUTHIDENTITY struct {
	// User is a pointer to a string containing the user name.
	User *uint16
	// UserLength is the length of the User string, excluding the terminating NULL character.
	UserLength uint32
	// Domain is a pointer to a string containing the domain or workgroup name.
	Domain *uint16
	// DomainLength is the length of the Domain string, excluding the terminating NULL character.
	DomainLength uint32
	// Password is a pointer to a string containing the user's password in the domain or workgroup.
	Password *uint16
	// PasswordLength is the length of the Password string, excluding the terminating NULL character.
	PasswordLength uint32
	// Flags indicates whether the strings are ANSI (SEC_WINNT_AUTH_IDENTITY_ANSI)
	// or Unicode (SEC_WINNT_AUTH_IDENTITY_UNICODE).
	Flags uint32
}

type COAUTHINFO struct {
	DwAuthnSvc           uint32
	DwAuthzSvc           uint32
	PwszServerPrincName  *uint16
	DwAuthnLevel         uint32
	DwImpersonationLevel uint32
	PAuthIdentityData    *COAUTHIDENTITY
	DwCapabilities       uint32
}

type COSERVERINFO struct {
	DwReserved1 uint32
	PwszName    *uint16
	PAuthInfo   *COAUTHINFO
	DwReserved2 uint32
}

type MULTI_QI struct {
	PIID *windows.GUID
	PItf *IUnknown
	Hr   int32 // long
}

// CoCreateInstanceEx creates an instance of a specific class on a specific computer.
// It allows for remote object creation and requesting multiple interfaces at once.
//
// Example:
//
//	err := com.CoCreateInstanceEx(clsid, nil, com.CLSCTX_REMOTE_SERVER, &serverInfo, 1, &results)
func CoCreateInstanceEx(Clsid *windows.GUID, punkOuter *IUnknown, dwClsCtx CLSCTX, pServerInfo *COSERVERINFO, dwCount uint32, pResults *MULTI_QI) (ret error) {
	r0, _, _ := syscall.SyscallN(procCoCreateInstanceEx.Addr(), uintptr(unsafe.Pointer(Clsid)), uintptr(unsafe.Pointer(punkOuter)), uintptr(dwClsCtx), uintptr(unsafe.Pointer(pServerInfo)), uintptr(dwCount), uintptr(unsafe.Pointer(pResults)))
	if r0 != 0 {
		ret = syscall.Errno(r0)
	}
	return
}

// VariantClear clears a VARIANT, releasing any resources it holds (like BSTRs or SafeArrays).
//
// Example:
//
//	var v com.VARIANT
//	// ... use v ...
//	com.VariantClear(&v)
func VariantClear(pvarg *VARIANT) (err error) {
	r0, _, _ := syscall.SyscallN(procVariantClear.Addr(), uintptr(unsafe.Pointer(pvarg)))
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func safeArrayGetVarType(safeArray *SafeArray) (varType uint16, err error) {
	r0, _, _ := syscall.SyscallN(procSafeArrayGetVarType.Addr(), uintptr(unsafe.Pointer(safeArray)), uintptr(unsafe.Pointer(&varType)))
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func safeArrayGetLBound(safeArray *SafeArray, dimension uint32) (lowerBound int32, err error) {
	r0, _, _ := syscall.SyscallN(
		procSafeArrayGetLBound.Addr(),
		uintptr(unsafe.Pointer(safeArray)),
		uintptr(dimension),
		uintptr(unsafe.Pointer(&lowerBound)),
	)
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func safeArrayGetUBound(safeArray *SafeArray, dimension uint32) (upperBound int32, err error) {
	r0, _, _ := syscall.SyscallN(
		procSafeArrayGetUBound.Addr(),
		uintptr(unsafe.Pointer(safeArray)),
		uintptr(dimension),
		uintptr(unsafe.Pointer(&upperBound)),
	)
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func safeArrayGetElement(safeArray *SafeArray, index int32, pv unsafe.Pointer) (err error) {
	r0, _, _ := syscall.SyscallN(
		procSafeArrayGetElement.Addr(),
		uintptr(unsafe.Pointer(safeArray)),
		uintptr(unsafe.Pointer(&index)),
		uintptr(pv))
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
	}
	return
}

// SysAllocStringLen allocates a new BSTR from a Go string.
// The returned pointer must eventually be freed with SysFreeString.
//
// Example:
//
//	bstr := com.SysAllocStringLen("Hello")
//	defer com.SysFreeString(bstr)
func SysAllocStringLen(v string) *uint16 {
	u := windows.StringToUTF16(v)
	r0, _, _ := syscall.SyscallN(
		procSysAllocStringLen.Addr(),
		uintptr(unsafe.Pointer(&u[0])),
		uintptr(len(u)-1),
	)
	return (*uint16)(unsafe.Pointer(r0))
}

func safeArrayCreateVector(variantType VT, lowerBound int32, length uint32) (safearray *SafeArray, err error) {
	r0, _, err := syscall.SyscallN(
		procSafeArrayCreateVector.Addr(),
		uintptr(variantType),
		uintptr(lowerBound),
		uintptr(length),
	)
	p0 := unsafe.Pointer(r0)
	if p0 == nil {
		if !errors.Is(err, windows.ERROR_SUCCESS) {
			return nil, err
		}
		return nil, syscall.EINVAL
	}
	return (*SafeArray)(p0), nil
}

func safeArrayPutElement(safearray *SafeArray, index int64, element uintptr) (err error) {
	r0, _, _ := syscall.SyscallN(
		procSafeArrayPutElement.Addr(),
		uintptr(unsafe.Pointer(safearray)),
		uintptr(unsafe.Pointer(&index)),
		element,
	)
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func SysFreeString(v *uint16) (err error) {
	r0, _, _ := syscall.SyscallN(
		procSysFreeString.Addr(),
		uintptr(unsafe.Pointer(v)),
	)
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

// MakeCOMObjectEx creates a COM object on a specified computer and returns its IUnknown interface.
// It simplifies the process of creating objects, especially on remote hosts.
//
// Example:
//
//	punk, err := com.MakeCOMObjectEx("remote-pc", com.CLSCTX_REMOTE_SERVER, clsid, iid)
func MakeCOMObjectEx(hostname string, serverLocation CLSCTX, requestedClass *windows.GUID, requestedInterface *windows.GUID) (*IUnknown, error) {
	reqInterface := MULTI_QI{
		PIID: requestedInterface,
		PItf: nil,
		Hr:   0,
	}
	var serverInfoPtr *COSERVERINFO = nil
	if serverLocation != CLSCTX_LOCAL_SERVER {
		serverInfoPtr = &COSERVERINFO{
			PwszName: windows.StringToUTF16Ptr(hostname),
		}
	}
	err := CoCreateInstanceEx(requestedClass, nil, serverLocation, serverInfoPtr, 1, &reqInterface)
	if err != nil {
		return nil, err
	}
	if reqInterface.Hr != 0 {
		return nil, syscall.Errno(reqInterface.Hr)
	}
	return reqInterface.PItf, nil
}

func IsLocal(host string) bool {
	if host == "" || host == "localhost" || host == "127.0.0.1" {
		return true
	}
	name, err := windows.ComputerName()
	if err != nil {
		return false
	}
	return strings.ToLower(name) == strings.ToLower(host)
}

// Initialize initializes the COM library on the current thread and sets the concurrency model to COINIT_MULTITHREADED.
// It also initializes COM security with default settings.
// This should be called once per application or thread before using any COM objects.
//
// Example:
//
//	if err := com.Initialize(); err != nil {
//	  log.Fatal(err)
//	}
//	defer com.Uninitialize()
func Initialize() error {
	config := DefaultInitConfig()
	return InitializeWithConfig(config)
}

type InitConfig struct {
	AuthLevel    uint32
	ImpLevel     uint32
	Capabilities uint32
}

func DefaultInitConfig() *InitConfig {
	return &InitConfig{
		AuthLevel:    RPC_C_AUTHN_LEVEL_NONE,
		ImpLevel:     RPC_C_IMP_LEVEL_IMPERSONATE,
		Capabilities: EOAC_NONE,
	}
}

func InitializeWithConfig(config *InitConfig) error {
	err := windows.CoInitializeEx(0, windows.COINIT_MULTITHREADED)
	if err != nil {
		return fmt.Errorf("call CoInitializeEx error: %s", err)
	}
	err = CoInitializeSecurity(config.AuthLevel, config.ImpLevel, config.Capabilities)
	if err != nil {
		Uninitialize()
		return fmt.Errorf("call CoInitializeSecurity error: %s", err)
	}
	return nil
}

// Uninitialize closes the COM library on the current thread.
// It should be called after all COM objects have been released and you are done with the COM library.
//
// Example:
//
//	defer com.Uninitialize()
func Uninitialize() {
	windows.CoUninitialize()
}

func IsEqualGUID(guid1 *windows.GUID, guid2 *windows.GUID) bool {
	return guid1.Data1 == guid2.Data1 &&
		guid1.Data2 == guid2.Data2 &&
		guid1.Data3 == guid2.Data3 &&
		guid1.Data4[0] == guid2.Data4[0] &&
		guid1.Data4[1] == guid2.Data4[1] &&
		guid1.Data4[2] == guid2.Data4[2] &&
		guid1.Data4[3] == guid2.Data4[3] &&
		guid1.Data4[4] == guid2.Data4[4] &&
		guid1.Data4[5] == guid2.Data4[5] &&
		guid1.Data4[6] == guid2.Data4[6] &&
		guid1.Data4[7] == guid2.Data4[7]
}

func CoInitializeSecurity(authnLevel, impLevel, capabilities uint32) (err error) {
	cAuthSvc := int32(-1)
	r0, _, _ := procCoInitializeSecurity.Call(
		uintptr(0),
		uintptr(cAuthSvc),
		uintptr(0),
		uintptr(0),
		uintptr(authnLevel),
		uintptr(impLevel),
		uintptr(0),
		uintptr(capabilities),
		uintptr(0))
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}
