//go:build windows

package opcda

import (
	"github.com/wends155/opcda/com"

	"golang.org/x/sys/windows"
)

// IID_CATID_OPCDAServer10 is the CATID for OPC DA 1.0 servers.
var IID_CATID_OPCDAServer10 = windows.GUID{
	Data1: 0x63D5F430,
	Data2: 0xCFE4,
	Data3: 0x11d1,
	Data4: [8]byte{0xB2, 0xC8, 0x00, 0x60, 0x08, 0x3B, 0xA1, 0xFB},
}

// IID_CATID_OPCDAServer20 is the CATID for OPC DA 2.0 servers.
var IID_CATID_OPCDAServer20 = windows.GUID{
	Data1: 0x63D5F432,
	Data2: 0xCFE4,
	Data3: 0x11d1,
	Data4: [8]byte{0xB2, 0xC8, 0x00, 0x60, 0x08, 0x3B, 0xA1, 0xFB},
}

// IID_IOPCShutdown is the GUID for the IOPCShutdown interface.
var IID_IOPCShutdown = windows.GUID{
	Data1: 0xF31DFDE1,
	Data2: 0x07B6,
	Data3: 0x11d2,
	Data4: [8]byte{0xB2, 0xD8, 0x00, 0x60, 0x08, 0x3B, 0xA1, 0xFB},
}

const (
	// OPC_READABLE indicates that the item is readable.
	OPC_READABLE uint32 = 0x1

	// OPC_WRITEABLE indicates that the item is writable.
	OPC_WRITEABLE = 0x2
)

const (
	// OPC_DS_CACHE indicates that the data should be read from the cache.
	OPC_DS_CACHE com.OPCDATASOURCE = 1
	// OPC_DS_DEVICE indicates that the data should be read from the device.
	OPC_DS_DEVICE com.OPCDATASOURCE = OPC_DS_CACHE + 1
)

const (
	// OPC_NS_HIERARCHIAL indicates that the address space is hierarchical.
	OPC_NS_HIERARCHIAL com.OPCNAMESPACETYPE = 1
	// OPC_NS_FLAT indicates that the address space is flat.
	OPC_NS_FLAT com.OPCNAMESPACETYPE = OPC_NS_HIERARCHIAL + 1
)

const (
	// OPC_STATUS_RUNNING indicates that the server is running.
	OPC_STATUS_RUNNING com.OPCServerState = 1
	// OPC_STATUS_FAILED indicates that the server has failed.
	OPC_STATUS_FAILED = OPC_STATUS_RUNNING + 1
	// OPC_STATUS_NOCONFIG indicates that the server is not configured.
	OPC_STATUS_NOCONFIG = OPC_STATUS_FAILED + 1
	// OPC_STATUS_SUSPENDED indicates that the server is suspended.
	OPC_STATUS_SUSPENDED = OPC_STATUS_NOCONFIG + 1
	// OPC_STATUS_TEST indicates that the server is in test mode.
	OPC_STATUS_TEST = OPC_STATUS_SUSPENDED + 1
	// OPC_STATUS_COMM_FAULT indicates that the server has a communication fault.
	OPC_STATUS_COMM_FAULT = OPC_STATUS_TEST + 1
)

const (
	// OPC_BROWSE_UP indicates browsing up.
	OPC_BROWSE_UP com.OPCBROWSEDIRECTION = 1
	// OPC_BROWSE_DOWN indicates browsing down.
	OPC_BROWSE_DOWN com.OPCBROWSEDIRECTION = OPC_BROWSE_UP + 1
	// OPC_BROWSE_TO indicates browsing to a specific position.
	OPC_BROWSE_TO com.OPCBROWSEDIRECTION = OPC_BROWSE_DOWN + 1
)

const (
	// OPC_BRANCH indicates browsing branches.
	OPC_BRANCH com.OPCBROWSETYPE = 1
	// OPC_LEAF indicates browsing leaves.
	OPC_LEAF com.OPCBROWSETYPE = OPC_BRANCH + 1
	// OPC_FLAT indicates browsing flat.
	OPC_FLAT com.OPCBROWSETYPE = OPC_LEAF + 1
)

const (
	// OPC_ENUM_PRIVATE_CONNECTIONS indicates private connections.
	OPC_ENUM_PRIVATE_CONNECTIONS = 1
	// OPC_ENUM_PUBLIC_CONNECTIONS indicates public connections.
	OPC_ENUM_PUBLIC_CONNECTIONS = OPC_ENUM_PRIVATE_CONNECTIONS + 1
	// OPC_ENUM_ALL_CONNECTIONS indicates all connections.
	OPC_ENUM_ALL_CONNECTIONS = OPC_ENUM_PUBLIC_CONNECTIONS + 1
	// OPC_ENUM_PRIVATE indicates private enumerations.
	OPC_ENUM_PRIVATE = OPC_ENUM_ALL_CONNECTIONS + 1
	// OPC_ENUM_PUBLIC indicates public enumerations.
	OPC_ENUM_PUBLIC = OPC_ENUM_PRIVATE + 1
	// OPC_ENUM_ALL indicates all enumerations.
	OPC_ENUM_ALL = OPC_ENUM_PUBLIC + 1
)
