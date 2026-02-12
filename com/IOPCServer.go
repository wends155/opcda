//go:build windows

package com

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var IID_IOPCServer = windows.GUID{
	Data1: 0x39c13a4d,
	Data2: 0x011e,
	Data3: 0x11d0,
	Data4: [8]byte{0x96, 0x75, 0x00, 0x20, 0xaf, 0xd8, 0xad, 0xb3},
}

// IOPCServer is the main interface for an OPC server as defined in the OPC Data Access Custom Interface Standard.
// It provides methods to manage groups, query status, and remove groups.
type IOPCServer struct {
	// IUnknown is the underlying COM interface.
	*IUnknown
}

// IOPCServerVtbl is the virtual function table for the IOPCServer interface.
type IOPCServerVtbl struct {
	IUnknownVtbl
	// AddGroup adds a new OPC group to the server.
	AddGroup uintptr
	// GetErrorString retrieves the error string for a server-specific error code.
	GetErrorString uintptr
	// GetGroupByName retrieves a pointer to an existing group by its name.
	GetGroupByName uintptr
	// GetStatus retrieves the current status of the server.
	GetStatus uintptr
	// RemoveGroup removes a group from the server.
	RemoveGroup uintptr
	// CreateGroupEnumerator creates an enumerator for the groups owned by the server.
	CreateGroupEnumerator uintptr
}

func (v *IOPCServer) Vtbl() *IOPCServerVtbl {
	return (*IOPCServerVtbl)(unsafe.Pointer(v.IUnknown.LpVtbl))
}

// AddGroup adds a new OPC group to the server.
//
// Parameters:
//
//	szName: The name of the group.
//	bActive: Whether the group should be active upon creation.
//	dwRequestedUpdateRate: The desired update rate in milliseconds.
//	hClientGroup: A client-side handle for the group.
//	pTimeBias: Optional time bias.
//	pPercentDeadband: Optional deadband percentage.
//	dwLCID: The locale ID for the group.
//	riid: The interface ID requested for the group object (usually IID_IOPCItemMgt).
//
// Example:
//
//	hServerGroup, revisedRate, pUnk, err := server.AddGroup("Group1", true, 1000, 1, nil, nil, 0x800, &com.IID_IOPCItemMgt)
func (v *IOPCServer) AddGroup(
	szName string,
	bActive bool,
	dwRequestedUpdateRate uint32,
	hClientGroup uint32,
	pTimeBias *int32,
	pPercentDeadband *float32,
	dwLCID uint32,
	riid *windows.GUID,
) (phServerGroup uint32, pRevisedUpdateRate uint32, ppUnk *IUnknown, err error) {
	var pUnk *IUnknown
	var pName *uint16
	pName, err = syscall.UTF16PtrFromString(szName)
	if err != nil {
		return
	}
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().AddGroup,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(unsafe.Pointer(pName)),
		uintptr(BoolToComBOOL(bActive)),
		uintptr(dwRequestedUpdateRate),
		uintptr(hClientGroup),
		uintptr(unsafe.Pointer(pTimeBias)),
		uintptr(unsafe.Pointer(pPercentDeadband)),
		uintptr(dwLCID),
		uintptr(unsafe.Pointer(&phServerGroup)),
		uintptr(unsafe.Pointer(&pRevisedUpdateRate)),
		uintptr(unsafe.Pointer(riid)),
		uintptr(unsafe.Pointer(&pUnk)),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	ppUnk = pUnk
	return
}

func BoolToComBOOL(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// OPCServerState represents the current state of the OPC server.
type OPCServerState uint32

// OPCSERVERSTATUS contains the current status, version information, and vendor data of an OPC server.
type OPCSERVERSTATUS struct {
	// FtStartTime is the time the server started.
	FtStartTime windows.Filetime
	// FtCurrentTime is the current time as seen by the server.
	FtCurrentTime windows.Filetime
	// FtLastUpdateTime is the last time the server was updated.
	FtLastUpdateTime windows.Filetime
	// DwServerState is the current state of the server.
	DwServerState OPCServerState
	// DwGroupCount is the number of groups currently defined in the server.
	DwGroupCount uint32
	// DwBandWidth is a measure of the server's current bandwidth usage.
	DwBandWidth uint32
	// WMajorVersion is the major version number of the server.
	WMajorVersion uint16
	// WMinorVersion is the minor version number of the server.
	WMinorVersion uint16
	// WBuildNumber is the build number of the server.
	WBuildNumber uint16
	// WReserved is reserved for future use.
	WReserved uint16
	// SzVendorInfo is a string containing vendor-specific information.
	SzVendorInfo *uint16
}

// ServerStatus is a Go-friendly version of OPCSERVERSTATUS.
type ServerStatus struct {
	// StartTime is the time the server started.
	StartTime time.Time
	// CurrentTime is the current time as seen by the server.
	CurrentTime time.Time
	// LastUpdateTime is the last time the server was updated.
	LastUpdateTime time.Time
	// ServerState is the current state of the server.
	ServerState OPCServerState
	// GroupCount is the number of groups currently defined in the server.
	GroupCount uint32
	// BandWidth is a measure of the server's current bandwidth usage.
	BandWidth uint32
	// MajorVersion is the major version number of the server.
	MajorVersion uint16
	// MinorVersion is the minor version number of the server.
	MinorVersion uint16
	// BuildNumber is the build number of the server.
	BuildNumber uint16
	// Reserved is reserved for future use.
	Reserved uint16
	// VendorInfo is a string containing vendor-specific information.
	VendorInfo string
}

// GetStatus retrieves the current status of the OPC server.
//
// Example:
//
//	status, err := server.GetStatus()
//	if err == nil {
//	  fmt.Printf("Server State: %v\n", status.ServerState)
//	}
func (v *IOPCServer) GetStatus() (status *ServerStatus, err error) {
	var pStatus *OPCSERVERSTATUS
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().GetStatus,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(unsafe.Pointer(&pStatus)),
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	defer func() {
		if pStatus != nil {
			if pStatus.SzVendorInfo != nil {
				CoTaskMemFree(unsafe.Pointer(pStatus.SzVendorInfo))
			}
			CoTaskMemFree(unsafe.Pointer(pStatus))
		}
	}()
	status = &ServerStatus{
		StartTime:      time.Unix(0, pStatus.FtStartTime.Nanoseconds()),
		CurrentTime:    time.Unix(0, pStatus.FtCurrentTime.Nanoseconds()),
		LastUpdateTime: time.Unix(0, pStatus.FtLastUpdateTime.Nanoseconds()),
		ServerState:    pStatus.DwServerState,
		GroupCount:     pStatus.DwGroupCount,
		BandWidth:      pStatus.DwBandWidth,
		MajorVersion:   pStatus.WMajorVersion,
		MinorVersion:   pStatus.WMinorVersion,
		BuildNumber:    pStatus.WBuildNumber,
		Reserved:       pStatus.WReserved,
		VendorInfo:     windows.UTF16PtrToString(pStatus.SzVendorInfo),
	}
	return
}

// RemoveGroup removes an OPC group from the server.
//
// Example:
//
//	err := server.RemoveGroup(hServerGroup, false)
func (v *IOPCServer) RemoveGroup(hServerGroup uint32, bForce bool) (err error) {
	r0, _, _ := syscall.SyscallN(
		v.Vtbl().RemoveGroup,
		uintptr(unsafe.Pointer(v.IUnknown)),
		uintptr(hServerGroup),
		uintptr(BoolToComBOOL(bForce)),
		0,
	)
	if int32(r0) < 0 {
		err = syscall.Errno(r0)
		return
	}
	return
}
