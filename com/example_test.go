//go:build windows

package com_test

import (
	"fmt"
	"log"

	"github.com/wends155/opcda/com"
)

func ExampleInitialize() {
	// Initialize the COM library.
	err := com.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize COM: %v", err)
	}
	defer com.Uninitialize()

	fmt.Println("COM initialized")
	// Output: COM initialized
}

func ExampleNewVariant() {
	// Create a new VARIANT with an integer value.
	v, err := com.NewVariant(int32(123))
	if err != nil {
		log.Fatal(err)
	}
	defer v.Clear()

	val, _ := v.Variant.Value()
	fmt.Printf("VT: %d, Value: %v\n", v.Variant.VT, val)
	// Output: VT: 3, Value: 123
}

func ExampleMakeCOMObjectEx() {
	// Example of creating a COM object (requires a valid CLSID and IID).
	/*
		clsid, _ := com.CLSIDFromProgID("Matrikon.OPC.Simulation.1")
		serverUnk, err := com.MakeCOMObjectEx("localhost", com.CLSCTX_LOCAL_SERVER, clsid, &com.IID_IOPCServer)
		if err != nil {
			log.Fatal(err)
		}
		server := &com.IOPCServer{IUnknown: serverUnk}
		defer server.Release()
	*/
}

func ExampleIOPCServer_AddGroup() {
	// This is a conceptual example as it requires a live OPC server.
	/*
		server, _ := GetOPCServer(...)
		hServerGroup, revisedRate, pUnk, err := server.AddGroup("Group1", true, 1000, 1, nil, nil, 0x800, &com.IID_IOPCItemMgt)
		if err == nil {
			fmt.Println("Group added")
			mgt := &com.IOPCItemMgt{IUnknown: pUnk}
			defer mgt.Release()
		}
	*/
}

func ExampleIOPCItemMgt_AddItems() {
	// Conceptual example for adding items to a group.
	/*
		mgt, _ := GetItemMgt(...)
		items := []com.TagOPCITEMDEF{
			{
				SzItemID: com.SysAllocStringLen("Random.Int4"),
				BActive:  1,
				HClient:  1,
			},
		}
		results, errors, err := mgt.AddItems(items)
		if err == nil {
			for i, res := range results {
				if errors[i] >= 0 {
					fmt.Printf("Item added with server handle: %d\n", res.Server)
				}
			}
		}
	*/
}
