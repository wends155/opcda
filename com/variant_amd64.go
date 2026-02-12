//go:build windows

package com

// VARIANT is the fundamental OLE Automation data type.
// It is a union-like structure that can hold many different types of data.
type VARIANT struct {
	// VT is the variant type.
	VT VT //  2
	// wReserved1 is reserved for future use.
	wReserved1 uint16 //  4
	// wReserved2 is reserved for future use.
	wReserved2 uint16 //  6
	// wReserved3 is reserved for future use.
	wReserved3 uint16 //  8
	// Val is a 64-bit value that holds the variant data.
	Val int64 // 16
	// _ is padding to ensure correct alignment on 64-bit systems.
	_ [8]byte // 24
}
