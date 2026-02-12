//go:build windows

package com

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// SafeArray represents an OLE automation array.
// It is a multidimensional array that carries its own bounds and dimensions.
// SafeArray represents an OLE automation array.
// It is a multidimensional array that carries its own bounds and dimensions.
type SafeArray struct {
	// Dimensions is the number of dimensions in the array.
	Dimensions uint16
	// FeaturesFlag is a set of flags describing the array.
	FeaturesFlag uint16
	// ElementsSize is the size of an array element in bytes.
	ElementsSize uint32
	// LocksAmount is the number of times the array has been locked.
	LocksAmount uint32
	// Data is a pointer to the array data.
	Data uint32
	// Bounds is a descriptor for each dimension of the array.
	Bounds [16]byte
}

// SafeArrayBound represents the bounds of one dimension of a SafeArray.
type SafeArrayBound struct {
	// Elements is the number of elements in the dimension.
	Elements uint32
	// LowerBound is the lower bound of the dimension (usually 0 or 1).
	LowerBound int32
}

// ToValueArray converts the SafeArray to a Go slice of values.
// It handles various VT types and returns an interface{} containing the resulting slice.
//
// Example:
//
//	slice, err := sa.ToValueArray()
//	if err == nil {
//	  fmt.Println(slice.([]float32))
//	}
//
//gocyclo:ignore
func (s *SafeArray) ToValueArray() (interface{}, error) {
	var err error
	totalElements, _ := s.TotalElements(0)
	vt, _ := safeArrayGetVarType(s)

	switch VT(vt) {
	case VT_BOOL:
		values := make([]bool, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v int16
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = (v & 0xff) != 0
		}
		return values, nil
	case VT_I1:
		values := make([]int8, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v int8
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_I2:
		values := make([]int16, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v int16
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_I4:
		values := make([]int32, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v int32
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_I8:
		values := make([]int64, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v int64
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_UI1:
		values := make([]uint8, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v uint8
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_UI2:
		values := make([]uint16, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v uint16
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_UI4:
		values := make([]uint32, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v uint32
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_UI8:
		values := make([]uint64, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v uint64
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_INT:
		values := make([]int, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v int
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_UINT:
		values := make([]uint, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v uint
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_R4:
		values := make([]float32, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v float32
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_R8:
		values := make([]float64, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v float64
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			values[i] = v
		}
		return values, nil
	case VT_BSTR:
		values := make([]string, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var element *uint16
			err = safeArrayGetElement(s, i, unsafe.Pointer(&element))
			if err != nil {
				return nil, err
			}
			values[i] = windows.UTF16PtrToString(element)
			SysFreeString(element)
		}
		return values, nil
	case VT_DATE:
		values := make([]time.Time, totalElements)
		for i := int32(0); i < totalElements; i++ {
			var v uint64
			err = safeArrayGetElement(s, i, unsafe.Pointer(&v))
			if err != nil {
				return nil, err
			}
			date, err := GetVariantDate(v)
			if err != nil {
				return nil, err
			}
			values[i] = date
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unknown value type %x", VT(vt))
	}
}

func (s *SafeArray) TotalElements(index uint32) (totalElements int32, err error) {
	if index < 1 {
		index = 1
	}

	// Get array bounds
	var LowerBounds int32
	var UpperBounds int32

	LowerBounds, err = safeArrayGetLBound(s, index)
	if err != nil {
		return
	}

	UpperBounds, err = safeArrayGetUBound(s, index)
	if err != nil {
		return
	}

	totalElements = UpperBounds - LowerBounds + 1
	return
}
