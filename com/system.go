//go:build windows

package com

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// IConnectionPointVtbl is the virtual function table for the IConnectionPoint interface.
type IConnectionPointVtbl struct {
	IUnknownVtbl
	// GetConnectionInterface retrieves the IID of the outgoing interface managed by this connection point.
	GetConnectionInterface uintptr
	// GetConnectionPointContainer retrieves the IConnectionPointContainer object that conceptually owns this connection point.
	GetConnectionPointContainer uintptr
	// Advise establishes a connection between the connection point object and the client's sink object.
	Advise uintptr
	// Unadvise terminates an advisory connection previously established through Advise.
	Unadvise uintptr
	// EnumConnections creates an enumerator object to iterate through the current connections for this connection point.
	EnumConnections uintptr
}

// IConnectionPoint supports connection points for outgoing interfaces (events/callbacks).
type IConnectionPoint struct {
	// IUnknown is the underlying COM interface.
	*IUnknown
}

func (p *IConnectionPoint) Vtbl() *IConnectionPointVtbl {
	return (*IConnectionPointVtbl)(unsafe.Pointer(p.IUnknown.LpVtbl))
}

func (p *IConnectionPoint) Advise(pUnkSink *IUnknown) (cookie uint32, err error) {
	r0, _, _ := syscall.SyscallN(p.Vtbl().Advise, uintptr(unsafe.Pointer(p.IUnknown)), uintptr(unsafe.Pointer(pUnkSink)), uintptr(unsafe.Pointer(&cookie)))
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
	}
	return
}

func (p *IConnectionPoint) Unadvise(dwCookie uint32) error {
	r0, _, _ := syscall.SyscallN(p.Vtbl().Unadvise, uintptr(unsafe.Pointer(p.IUnknown)), uintptr(dwCookie))
	if int32(r0) < 0 {
		return syscall.Errno(r0)
	}
	return nil
}

// B196B284-BAB4-101A-B69C-00AA00341D07
var IID_IConnectionPointContainer = windows.GUID{
	Data1: 0xB196B284,
	Data2: 0xBAB4,
	Data3: 0x101A,
	Data4: [8]byte{0xB6, 0x9C, 0x00, 0xAA, 0x00, 0x34, 0x1D, 0x07},
}

// IConnectionPointContainerVtbl is the virtual function table for the IConnectionPointContainer interface.
type IConnectionPointContainerVtbl struct {
	IUnknownVtbl
	// EnumConnectionPoints creates an enumerator object to iterate through all the connection points supported in the connectable object.
	EnumConnectionPoints uintptr
	// FindConnectionPoint returns a pointer to the IConnectionPoint interface for a specific IID.
	FindConnectionPoint uintptr
}

// IConnectionPointContainer indicates that an object is connectable and provides access to its connection points.
type IConnectionPointContainer struct {
	// IUnknown is the underlying COM interface.
	*IUnknown
}

func (c *IConnectionPointContainer) Vtbl() *IConnectionPointContainerVtbl {
	return (*IConnectionPointContainerVtbl)(unsafe.Pointer(c.IUnknown.LpVtbl))
}

func (c *IConnectionPointContainer) FindConnectionPoint(riid *windows.GUID) (*IConnectionPoint, error) {
	var iUnknown *IUnknown
	r0, _, _ := syscall.SyscallN(
		c.Vtbl().FindConnectionPoint,
		uintptr(unsafe.Pointer(c.IUnknown)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(&iUnknown)),
	)
	if int32(r0) < 0 {
		return nil, syscall.Errno(r0)
	}
	return &IConnectionPoint{iUnknown}, nil
}
