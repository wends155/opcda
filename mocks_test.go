//go:build windows

package opcda

import (
	"unsafe"

	"github.com/wends155/opcda/com"
	"golang.org/x/sys/windows"
)

// mockServerProvider implements the serverProvider interface for testing.
type mockServerProvider struct {
	GetStatusFn                func() (*com.ServerStatus, error)
	GetErrorStringFn           func(errorCode uint32) (string, error)
	GetLocaleIDFn              func() (uint32, error)
	SetLocaleIDFn              func(localeID uint32) error
	SetClientNameFn            func(clientName string) error
	QueryAvailableLocaleIDsFn  func() ([]uint32, error)
	QueryAvailablePropertiesFn func(itemID string) ([]uint32, []string, []uint16, error)
	GetItemPropertiesFn        func(itemID string, propertyIDs []uint32) ([]interface{}, []int32, error)
	LookupItemIDsFn            func(itemID string, propertyIDs []uint32) ([]string, []int32, error)
	AddGroupFn                 func(name string, active bool, updateRate uint32, clientGroup uint32, timeBias *int32, deadband *float32, localeID uint32, iid *windows.GUID) (uint32, uint32, *com.IUnknown, error)
	RemoveGroupFn              func(serverGroup uint32, force bool) error
	ReleaseFn                  func()
	QueryInterfaceFn           func(iid *windows.GUID, ppv unsafe.Pointer) error
}

func (m *mockServerProvider) GetStatus() (*com.ServerStatus, error) {
	if m.GetStatusFn != nil {
		return m.GetStatusFn()
	}
	return &com.ServerStatus{}, nil
}

func (m *mockServerProvider) GetErrorString(errorCode uint32) (string, error) {
	if m.GetErrorStringFn != nil {
		return m.GetErrorStringFn(errorCode)
	}
	return "mock error", nil
}

func (m *mockServerProvider) GetLocaleID() (uint32, error) {
	if m.GetLocaleIDFn != nil {
		return m.GetLocaleIDFn()
	}
	return 0, nil
}

func (m *mockServerProvider) SetLocaleID(localeID uint32) error {
	if m.SetLocaleIDFn != nil {
		return m.SetLocaleIDFn(localeID)
	}
	return nil
}

func (m *mockServerProvider) SetClientName(clientName string) error {
	if m.SetClientNameFn != nil {
		return m.SetClientNameFn(clientName)
	}
	return nil
}

func (m *mockServerProvider) QueryAvailableLocaleIDs() ([]uint32, error) {
	if m.QueryAvailableLocaleIDsFn != nil {
		return m.QueryAvailableLocaleIDsFn()
	}
	return []uint32{1033}, nil
}

func (m *mockServerProvider) QueryAvailableProperties(itemID string) ([]uint32, []string, []uint16, error) {
	if m.QueryAvailablePropertiesFn != nil {
		return m.QueryAvailablePropertiesFn(itemID)
	}
	return nil, nil, nil, nil
}

func (m *mockServerProvider) GetItemProperties(itemID string, propertyIDs []uint32) ([]interface{}, []int32, error) {
	if m.GetItemPropertiesFn != nil {
		return m.GetItemPropertiesFn(itemID, propertyIDs)
	}
	return nil, nil, nil
}

func (m *mockServerProvider) LookupItemIDs(itemID string, propertyIDs []uint32) ([]string, []int32, error) {
	if m.LookupItemIDsFn != nil {
		return m.LookupItemIDsFn(itemID, propertyIDs)
	}
	return nil, nil, nil
}

func (m *mockServerProvider) AddGroup(name string, active bool, updateRate uint32, clientGroup uint32, timeBias *int32, deadband *float32, localeID uint32, iid *windows.GUID) (uint32, uint32, *com.IUnknown, error) {
	if m.AddGroupFn != nil {
		return m.AddGroupFn(name, active, updateRate, clientGroup, timeBias, deadband, localeID, iid)
	}
	return 0, updateRate, nil, nil
}

func (m *mockServerProvider) RemoveGroup(serverGroup uint32, force bool) error {
	if m.RemoveGroupFn != nil {
		return m.RemoveGroupFn(serverGroup, force)
	}
	return nil
}

func (m *mockServerProvider) Release() {
	if m.ReleaseFn != nil {
		m.ReleaseFn()
	}
}

func (m *mockServerProvider) QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error {
	if m.QueryInterfaceFn != nil {
		return m.QueryInterfaceFn(iid, ppv)
	}
	return nil
}

// mockGroupProvider implements the groupProvider interface for testing.
type mockGroupProvider struct {
	SetNameFn        func(name string) error
	GetStateFn       func() (uint32, bool, string, int32, float32, uint32, uint32, uint32, error)
	SetStateFn       func(pRequestedUpdateRate *uint32, pActive *int32, pTimeBias *int32, pPercentDeadband *float32, pLCID *uint32, phClientGroup *uint32) (uint32, error)
	SyncReadFn       func(source com.OPCDATASOURCE, serverHandles []uint32) ([]*com.ItemState, []int32, error)
	SyncWriteFn      func(serverHandles []uint32, values []com.VARIANT) ([]int32, error)
	AsyncReadFn      func(serverHandles []uint32, transactionID uint32) (uint32, []int32, error)
	AsyncWriteFn     func(serverHandles []uint32, values []com.VARIANT, transactionID uint32) (uint32, []int32, error)
	AsyncRefreshFn   func(source com.OPCDATASOURCE, transactionID uint32) (uint32, error)
	AsyncCancelFn    func(cancelID uint32) error
	QueryInterfaceFn func(iid *windows.GUID, ppv unsafe.Pointer) error
	ReleaseFn        func()
}

func (m *mockGroupProvider) SetName(name string) error {
	if m.SetNameFn != nil {
		return m.SetNameFn(name)
	}
	return nil
}

func (m *mockGroupProvider) GetState() (uint32, bool, string, int32, float32, uint32, uint32, uint32, error) {
	if m.GetStateFn != nil {
		return m.GetStateFn()
	}
	return 1000, true, "mock", 0, 0, 1033, 0, 0, nil
}

func (m *mockGroupProvider) SetState(pRequestedUpdateRate *uint32, pActive *int32, pTimeBias *int32, pPercentDeadband *float32, pLCID *uint32, phClientGroup *uint32) (uint32, error) {
	if m.SetStateFn != nil {
		return m.SetStateFn(pRequestedUpdateRate, pActive, pTimeBias, pPercentDeadband, pLCID, phClientGroup)
	}
	if pRequestedUpdateRate != nil {
		return *pRequestedUpdateRate, nil
	}
	return 1000, nil
}

func (m *mockGroupProvider) SyncRead(source com.OPCDATASOURCE, serverHandles []uint32) ([]*com.ItemState, []int32, error) {
	if m.SyncReadFn != nil {
		return m.SyncReadFn(source, serverHandles)
	}
	return make([]*com.ItemState, len(serverHandles)), make([]int32, len(serverHandles)), nil
}

func (m *mockGroupProvider) SyncWrite(serverHandles []uint32, values []com.VARIANT) ([]int32, error) {
	if m.SyncWriteFn != nil {
		return m.SyncWriteFn(serverHandles, values)
	}
	return make([]int32, len(serverHandles)), nil
}

func (m *mockGroupProvider) AsyncRead(serverHandles []uint32, transactionID uint32) (uint32, []int32, error) {
	if m.AsyncReadFn != nil {
		return m.AsyncReadFn(serverHandles, transactionID)
	}
	return 1, make([]int32, len(serverHandles)), nil
}

func (m *mockGroupProvider) AsyncWrite(serverHandles []uint32, values []com.VARIANT, transactionID uint32) (uint32, []int32, error) {
	if m.AsyncWriteFn != nil {
		return m.AsyncWriteFn(serverHandles, values, transactionID)
	}
	return 1, make([]int32, len(serverHandles)), nil
}

func (m *mockGroupProvider) AsyncRefresh(source com.OPCDATASOURCE, transactionID uint32) (uint32, error) {
	if m.AsyncRefreshFn != nil {
		return m.AsyncRefreshFn(source, transactionID)
	}
	return 1, nil
}

func (m *mockGroupProvider) AsyncCancel(cancelID uint32) error {
	if m.AsyncCancelFn != nil {
		return m.AsyncCancelFn(cancelID)
	}
	return nil
}

func (m *mockGroupProvider) QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error {
	if m.QueryInterfaceFn != nil {
		return m.QueryInterfaceFn(iid, ppv)
	}
	return nil
}

func (m *mockGroupProvider) Release() {
	if m.ReleaseFn != nil {
		m.ReleaseFn()
	}
}

// mockItemMgtProvider implements the itemMgtProvider interface for testing.
type mockItemMgtProvider struct {
	AddItemsFn         func(items []com.TagOPCITEMDEF) ([]com.TagOPCITEMRESULTStruct, []int32, error)
	ValidateItemsFn    func(items []com.TagOPCITEMDEF, bBlob bool) ([]com.TagOPCITEMRESULTStruct, []int32, error)
	RemoveItemsFn      func(serverHandles []uint32) ([]int32, error)
	SetActiveStateFn   func(serverHandles []uint32, bActive bool) ([]int32, error)
	SetClientHandlesFn func(serverHandles []uint32, clientHandles []uint32) ([]int32, error)
	SetDatatypesFn     func(serverHandles []uint32, requestedDataTypes []com.VT) ([]int32, error)
	ReleaseFn          func()
}

func (m *mockItemMgtProvider) AddItems(items []com.TagOPCITEMDEF) ([]com.TagOPCITEMRESULTStruct, []int32, error) {
	if m.AddItemsFn != nil {
		return m.AddItemsFn(items)
	}
	return make([]com.TagOPCITEMRESULTStruct, len(items)), make([]int32, len(items)), nil
}

func (m *mockItemMgtProvider) ValidateItems(items []com.TagOPCITEMDEF, bBlob bool) ([]com.TagOPCITEMRESULTStruct, []int32, error) {
	if m.ValidateItemsFn != nil {
		return m.ValidateItemsFn(items, bBlob)
	}
	return make([]com.TagOPCITEMRESULTStruct, len(items)), make([]int32, len(items)), nil
}

func (m *mockItemMgtProvider) RemoveItems(serverHandles []uint32) ([]int32, error) {
	if m.RemoveItemsFn != nil {
		return m.RemoveItemsFn(serverHandles)
	}
	return make([]int32, len(serverHandles)), nil
}

func (m *mockItemMgtProvider) SetActiveState(serverHandles []uint32, bActive bool) ([]int32, error) {
	if m.SetActiveStateFn != nil {
		return m.SetActiveStateFn(serverHandles, bActive)
	}
	return make([]int32, len(serverHandles)), nil
}

func (m *mockItemMgtProvider) SetClientHandles(serverHandles []uint32, clientHandles []uint32) ([]int32, error) {
	if m.SetClientHandlesFn != nil {
		return m.SetClientHandlesFn(serverHandles, clientHandles)
	}
	return make([]int32, len(serverHandles)), nil
}

func (m *mockItemMgtProvider) SetDatatypes(serverHandles []uint32, requestedDataTypes []com.VT) ([]int32, error) {
	if m.SetDatatypesFn != nil {
		return m.SetDatatypesFn(serverHandles, requestedDataTypes)
	}
	return make([]int32, len(serverHandles)), nil
}

func (m *mockItemMgtProvider) Release() {
	if m.ReleaseFn != nil {
		m.ReleaseFn()
	}
}
