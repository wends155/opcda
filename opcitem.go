//go:build windows

package opcda

import (
	"errors"
	"sync"
	"time"

	"github.com/wends155/opcda/com"
)

type OPCItem struct {
	itemMgtProvider itemMgtProvider
	groupProvider   groupProvider
	provider        serverProvider
	sync.RWMutex
	value             interface{}
	quality           uint16
	timestamp         time.Time
	serverHandle      uint32
	clientHandle      uint32
	tag               string
	accessPath        string
	accessRights      uint32
	isActive          bool
	requestedDataType com.VT
	nativeDataType    com.VT
	parent            *OPCItems
}

// GetParent Returns reference to the parent OPCItems object.
func (i *OPCItem) GetParent() *OPCItems {
	if i == nil {
		return nil
	}
	return i.parent
}

// GetClientHandle get the client handle for the item.
func (i *OPCItem) GetClientHandle() uint32 {
	if i == nil {
		return 0
	}
	return i.clientHandle
}

// SetClientHandle set the client handle for the item.
func (i *OPCItem) SetClientHandle(clientHandle uint32) error {
	if i == nil || i.itemMgtProvider == nil {
		return errors.New("uninitialized item")
	}
	errs, err := i.itemMgtProvider.SetClientHandles([]uint32{i.serverHandle}, []uint32{clientHandle})
	if err != nil {
		return err
	}
	if errs[0] < 0 {
		return i.getError(errs[0])
	}
	i.Lock()
	i.clientHandle = clientHandle
	i.Unlock()
	return nil
}

// GetServerHandle get the server handle for the item.
func (i *OPCItem) GetServerHandle() uint32 {
	if i == nil {
		return 0
	}
	return i.serverHandle
}

// GetAccessPath get the access path for the item.
func (i *OPCItem) GetAccessPath() string {
	if i == nil {
		return ""
	}
	return i.accessPath
}

// GetAccessRights get the access rights for the item.
func (i *OPCItem) GetAccessRights() uint32 {
	if i == nil {
		return 0
	}
	return i.accessRights
}

// GetItemID get the item ID for the item.
func (i *OPCItem) GetItemID() string {
	if i == nil {
		return ""
	}
	return i.tag
}

// GetIsActive get the active state for the item.
func (i *OPCItem) GetIsActive() bool {
	if i == nil {
		return false
	}
	i.RLock()
	defer i.RUnlock()
	return i.isActive
}

// GetRequestedDataType get the requested data type for the item.
func (i *OPCItem) GetRequestedDataType() com.VT {
	if i == nil {
		return com.VT_EMPTY
	}
	i.RLock()
	defer i.RUnlock()
	return i.requestedDataType
}

// SetRequestedDataType set the requested data type for the item.
func (i *OPCItem) SetRequestedDataType(requestedDataType com.VT) error {
	if i == nil || i.itemMgtProvider == nil {
		return errors.New("uninitialized item")
	}
	errs, err := i.itemMgtProvider.SetDatatypes([]uint32{i.serverHandle}, []com.VT{requestedDataType})
	if err != nil {
		return err
	}
	if errs[0] < 0 {
		return i.getError(errs[0])
	}
	i.Lock()
	i.requestedDataType = requestedDataType
	i.Unlock()
	return nil
}

// SetIsActive set the active state for the item.
func (i *OPCItem) SetIsActive(isActive bool) error {
	if i == nil || i.itemMgtProvider == nil {
		return errors.New("uninitialized item")
	}
	errs, err := i.itemMgtProvider.SetActiveState([]uint32{i.serverHandle}, isActive)
	if err != nil {
		return err
	}
	if errs[0] < 0 {
		return i.getError(errs[0])
	}
	i.Lock()
	i.isActive = isActive
	i.Unlock()
	return nil
}

// GetValue Returns the latest value read from the server
func (i *OPCItem) GetValue() interface{} {
	if i == nil {
		return nil
	}
	i.RLock()
	defer i.RUnlock()
	return i.value
}

// GetQuality Returns the latest quality read from the server
func (i *OPCItem) GetQuality() uint16 {
	if i == nil {
		return 0
	}
	i.RLock()
	defer i.RUnlock()
	return i.quality
}

// GetTimestamp Returns the latest timestamp read from the server
func (i *OPCItem) GetTimestamp() time.Time {
	if i == nil {
		return time.Time{}
	}
	i.RLock()
	defer i.RUnlock()
	return i.timestamp
}

// GetCanonicalDataType Returns the canonical data type for the item.
func (i *OPCItem) GetCanonicalDataType() com.VT {
	if i == nil {
		return com.VT_EMPTY
	}
	return i.nativeDataType
}

// GetEUType Returns the EU type for the item.
func (i *OPCItem) GetEUType() (int, error) {
	if i == nil || i.parent == nil || i.parent.parent == nil || i.parent.parent.parent == nil {
		return 0, errors.New("uninitialized item")
	}
	data, errs, err := i.parent.parent.parent.parent.GetItemProperties(i.tag, []uint32{7})
	if err != nil {
		return 0, err
	}
	if errs[0] != nil {
		return 0, errs[0]
	}
	return (int)(data[0].(int32)), nil
}

// GetEUInfo Returns the EU info for the item.
func (i *OPCItem) GetEUInfo() (interface{}, error) {
	if i == nil {
		return nil, errors.New("uninitialized item")
	}
	euType, err := i.GetEUType()
	if err != nil {
		return nil, err
	}
	if euType == 0 {
		return nil, nil
	}
	if euType > 2 {
		return nil, errors.New("not valid")
	}
	data, errs, err := i.parent.parent.parent.parent.GetItemProperties(i.tag, []uint32{8})
	if err != nil {
		return nil, err
	}
	if errs[0] != nil {
		return nil, errs[0]
	}
	return data[0], nil
}

func NewOPCItem(
	parent *OPCItems,
	tag string,
	result com.TagOPCITEMRESULTStruct,
	clientHandle uint32,
	accessPath string,
	isActive bool,
) *OPCItem {
	return &OPCItem{
		itemMgtProvider: parent.itemMgtProvider,
		groupProvider:   parent.parent.groupProvider,
		provider:        parent.provider,
		parent:          parent,
		tag:             tag,
		accessPath:      accessPath,
		serverHandle:    result.Server,
		clientHandle:    clientHandle,
		accessRights:    result.AccessRights,
		nativeDataType:  com.VT(result.NativeType),
		isActive:        isActive,
	}
}

// Read reads the value, quality and timestamp for the item.
func (i *OPCItem) Read(source com.OPCDATASOURCE) (interface{}, uint16, time.Time, error) {
	if i == nil || i.groupProvider == nil {
		return nil, 0, time.Time{}, errors.New("uninitialized item")
	}
	values, errs, err := i.groupProvider.SyncRead(source, []uint32{i.serverHandle})
	if err != nil {
		return nil, 0, time.Time{}, err
	}
	if errs[0] < 0 {
		return nil, 0, time.Time{}, i.getError(errs[0])
	}
	val := values[0].Value
	qual := values[0].Quality
	ts := values[0].Timestamp

	i.Lock()
	i.value = val
	i.quality = qual
	i.timestamp = ts
	i.Unlock()
	return val, qual, ts, nil
}

// Write writes a value to the item.
func (i *OPCItem) Write(value interface{}) error {
	if i == nil || i.groupProvider == nil {
		return errors.New("uninitialized item")
	}
	variant, err := com.NewVariant(value)
	if err != nil {
		return err
	}
	defer variant.Clear()
	errs, err := i.groupProvider.SyncWrite([]uint32{i.serverHandle}, []com.VARIANT{*variant.Variant})
	if err != nil {
		return err
	}
	if errs[0] < 0 {
		return i.getError(errs[0])
	}
	return nil
}

func (i *OPCItem) getError(errorCode int32) error {
	if i == nil || i.provider == nil {
		return &OPCError{ErrorCode: errorCode, ErrorMessage: "uninitialized common interface"}
	}
	errStr, _ := i.provider.GetErrorString(uint32(errorCode))
	return &OPCError{
		ErrorCode:    errorCode,
		ErrorMessage: errStr,
	}
}

// Release Releases the OPCItem object
func (i *OPCItem) Release() {
	// No interfaces owned by OPCItem explicitly for now
}
