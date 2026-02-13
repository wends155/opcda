//go:build windows

package opcda

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/wends155/opcda/com"

	"golang.org/x/sys/windows"
)

// itemMgtProvider defines the internal contract for managing OPC items.
// It abstracts the underlying COM implementation to allow for mocking and testing.
type itemMgtProvider interface {
	// AddItems adds items to the group.
	AddItems(items []com.TagOPCITEMDEF) ([]com.TagOPCITEMRESULTStruct, []int32, error)
	// ValidateItems validates items without adding them.
	ValidateItems(items []com.TagOPCITEMDEF, bBlob bool) ([]com.TagOPCITEMRESULTStruct, []int32, error)
	// RemoveItems removes items from the group.
	RemoveItems(serverHandles []uint32) ([]int32, error)
	// SetActiveState sets the active state for the specified items.
	SetActiveState(serverHandles []uint32, bActive bool) ([]int32, error)
	// SetClientHandles sets the client handles for the specified items.
	SetClientHandles(serverHandles []uint32, clientHandles []uint32) ([]int32, error)
	// SetDatatypes sets the requested data types for the specified items.
	SetDatatypes(serverHandles []uint32, requestedDataTypes []com.VT) ([]int32, error)
	// Release releases the COM resources associated with the provider.
	Release()
}

// comItemMgtProvider is the concrete implementation of itemMgtProvider using COM.
type comItemMgtProvider struct {
	itemMgt *com.IOPCItemMgt
}

// AddItems adds items to the group.
func (p *comItemMgtProvider) AddItems(items []com.TagOPCITEMDEF) ([]com.TagOPCITEMRESULTStruct, []int32, error) {
	return p.itemMgt.AddItems(items)
}

// ValidateItems validates items without adding them.
func (p *comItemMgtProvider) ValidateItems(items []com.TagOPCITEMDEF, bBlob bool) ([]com.TagOPCITEMRESULTStruct, []int32, error) {
	return p.itemMgt.ValidateItems(items, bBlob)
}

// RemoveItems removes items from the group.
func (p *comItemMgtProvider) RemoveItems(serverHandles []uint32) ([]int32, error) {
	return p.itemMgt.RemoveItems(serverHandles)
}

// SetActiveState sets the active state for the specified items.
func (p *comItemMgtProvider) SetActiveState(serverHandles []uint32, bActive bool) ([]int32, error) {
	return p.itemMgt.SetActiveState(serverHandles, bActive)
}

// SetClientHandles sets the client handles for the specified items.
func (p *comItemMgtProvider) SetClientHandles(serverHandles []uint32, clientHandles []uint32) ([]int32, error) {
	return p.itemMgt.SetClientHandles(serverHandles, clientHandles)
}

// SetDatatypes sets the requested data types for the specified items.
func (p *comItemMgtProvider) SetDatatypes(serverHandles []uint32, requestedDataTypes []com.VT) ([]int32, error) {
	return p.itemMgt.SetDatatypes(serverHandles, requestedDataTypes)
}

// Release releases the COM resources associated with the provider.
func (p *comItemMgtProvider) Release() {
	p.itemMgt.Release()
}

// OPCItems represents a collection of OPC items.
type OPCItems struct {
	itemMgtProvider          itemMgtProvider
	provider                 serverProvider
	parent                   *OPCGroup
	itemID                   uint32
	defaultRequestedDataType com.VT
	defaultAccessPath        string
	defaultActive            bool
	items                    []*OPCItem
	sync.RWMutex
}

// NewOPCItems creates a new OPCItems instance.
func NewOPCItems(
	parent *OPCGroup,
	itemMgt itemMgtProvider,
	provider serverProvider,
) *OPCItems {
	return &OPCItems{
		parent:                   parent,
		itemMgtProvider:          itemMgt,
		defaultRequestedDataType: com.VT_EMPTY,
		defaultAccessPath:        "",
		defaultActive:            true,
		provider:                 provider,
	}
}

// GetParent returns a reference to the parent OPCGroup object.
func (is *OPCItems) GetParent() *OPCGroup {
	if is == nil {
		return nil
	}
	return is.parent
}

// GetDefaultRequestedDataType returns the requested data type that will be used in calls to Add.
func (is *OPCItems) GetDefaultRequestedDataType() com.VT {
	if is == nil {
		return com.VT_EMPTY
	}
	return is.defaultRequestedDataType
}

// SetDefaultRequestedDataType sets the requested data type that will be used in calls to Add.
func (is *OPCItems) SetDefaultRequestedDataType(defaultRequestedDataType com.VT) {
	if is == nil {
		return
	}
	is.defaultRequestedDataType = defaultRequestedDataType
}

// GetDefaultAccessPath returns the default AccessPath that will be used in calls to Add.
func (is *OPCItems) GetDefaultAccessPath() string {
	if is == nil {
		return ""
	}
	return is.defaultAccessPath
}

// SetDefaultAccessPath sets the default AccessPath that will be used in calls to Add.
func (is *OPCItems) SetDefaultAccessPath(defaultAccessPath string) {
	if is == nil {
		return
	}
	is.Lock()
	defer is.Unlock()
	is.defaultAccessPath = defaultAccessPath
}

// GetDefaultActive returns the default active state for OPCItems created using Items.Add.
func (is *OPCItems) GetDefaultActive() bool {
	if is == nil {
		return false
	}
	return is.defaultActive
}

// SetDefaultActive sets the default active state for OPCItems created using Items.Add.
func (is *OPCItems) SetDefaultActive(defaultActive bool) {
	if is == nil {
		return
	}
	is.defaultActive = defaultActive
}

// GetCount returns the number of items in the collection.
func (is *OPCItems) GetCount() int {
	if is == nil {
		return 0
	}
	is.RLock()
	defer is.RUnlock()
	return len(is.items)
}

// Item returns the item by index.
func (is *OPCItems) Item(index int32) (*OPCItem, error) {
	if is == nil {
		return nil, errors.New("uninitialized items")
	}
	is.RLock()
	defer is.RUnlock()
	if index < 0 || index >= int32(len(is.items)) {
		return nil, errors.New("index out of range")
	}
	return is.items[index], nil
}

// ItemByName returns the item by name.
func (is *OPCItems) ItemByName(name string) (*OPCItem, error) {
	if is == nil {
		return nil, errors.New("uninitialized items")
	}
	is.RLock()
	defer is.RUnlock()
	for _, v := range is.items {
		if v.tag == name {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}

// GetOPCItem returns the OPCItem by serverHandle.
func (is *OPCItems) GetOPCItem(serverHandle uint32) (*OPCItem, error) {
	if is == nil {
		return nil, errors.New("uninitialized items")
	}
	is.RLock()
	defer is.RUnlock()
	for _, v := range is.items {
		if v.serverHandle == serverHandle {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}

// AddItem adds an item to the group.
func (is *OPCItems) AddItem(tag string) (*OPCItem, error) {
	if is == nil || is.itemMgtProvider == nil {
		return nil, errors.New("uninitialized items or failed group connection")
	}
	items, errs, err := is.AddItems([]string{tag})
	if err != nil {
		return nil, err
	}
	if errs[0] != nil {
		return nil, errs[0]
	}
	return items[0], nil
}

// AddItems adds multiple items to the collection.
func (is *OPCItems) AddItems(tags []string) ([]*OPCItem, []error, error) {
	if is == nil || is.itemMgtProvider == nil {
		return nil, nil, errors.New("uninitialized items or failed group connection")
	}
	is.Lock()
	defer is.Unlock()
	accessPath := is.defaultAccessPath
	active := is.defaultActive
	dt := is.defaultRequestedDataType
	items := is.createDefinitions(tags, accessPath, active, dt)
	results, errs, err := is.itemMgtProvider.AddItems(items)
	if err != nil {
		return nil, nil, err
	}
	var resultErrors = make([]error, len(tags))
	var opcItems = make([]*OPCItem, len(tags))
	for j := 0; j < len(tags); j++ {
		if errs[j] < 0 {
			resultErrors[j] = is.getError(errs[j])
		} else {
			item := NewOPCItem(is, tags[j], results[j], items[j].HClient, accessPath, active)
			opcItems[j] = item
			is.items = append(is.items, item)
		}
	}
	return opcItems, resultErrors, nil
}

// Remove removes an OPCItem from the collection.
func (is *OPCItems) Remove(serverHandles []uint32) {
	if is == nil {
		return
	}
	is.Lock()
	defer is.Unlock()
	toDelete := make(map[uint32]struct{}, len(serverHandles))
	for _, h := range serverHandles {
		toDelete[h] = struct{}{}
	}
	var newItems []*OPCItem
	var removedItems []*OPCItem
	var removedHandles []uint32
	for _, item := range is.items {
		if _, ok := toDelete[item.serverHandle]; ok {
			removedItems = append(removedItems, item)
			removedHandles = append(removedHandles, item.serverHandle)
			continue
		}
		newItems = append(newItems, item)
	}

	is.items = newItems

	if len(removedHandles) > 0 {
		if is.itemMgtProvider != nil {
			is.itemMgtProvider.RemoveItems(removedHandles)
		}
	}
	for _, it := range removedItems {
		it.Release()
	}
}

// Validate determines if one or more OPCItems could be successfully created via the Add method (but does not add them).
func (is *OPCItems) Validate(tags []string, requestedDataTypes *[]com.VT, accessPaths *[]string) ([]error, error) {
	if is == nil || is.itemMgtProvider == nil {
		return nil, errors.New("uninitialized items or failed group connection")
	}
	var definitions []com.TagOPCITEMDEF
	for i, v := range tags {
		cHandle := atomic.AddUint32(&is.itemID, 1)
		item := com.TagOPCITEMDEF{
			SzAccessPath: windows.StringToUTF16Ptr(""),
			SzItemID:     windows.StringToUTF16Ptr(v),
			BActive:      com.BoolToComBOOL(false),
			HClient:      cHandle,
			DwBlobSize:   0,
			PBlob:        nil,
			VtRequested:  uint16(is.defaultRequestedDataType),
		}
		if requestedDataTypes != nil {
			item.VtRequested = uint16((*requestedDataTypes)[i])
		}
		if accessPaths != nil {
			item.SzAccessPath = windows.StringToUTF16Ptr((*accessPaths)[i])
		}
		definitions = append(definitions, item)
	}
	_, errs, err := is.itemMgtProvider.ValidateItems(definitions, false)
	if err != nil {
		return nil, err
	}
	var resultErrors = make([]error, len(errs))
	for j := 0; j < len(errs); j++ {
		if errs[j] < 0 {
			resultErrors[j] = is.getError(errs[j])
		}
	}
	return resultErrors, nil
}

// SetActive allows activation and deactivation of individual OPCItems in the OPCItems collection.
func (is *OPCItems) SetActive(serverHandles []uint32, active bool) []error {
	if is == nil {
		return nil
	}
	resultErrors := make([]error, len(serverHandles))
	for i, handle := range serverHandles {
		item, err := is.GetOPCItem(handle)
		if err != nil {
			resultErrors[i] = err
			continue
		}
		err = item.SetIsActive(active)
		if err != nil {
			resultErrors[i] = err
		}
	}
	return resultErrors
}

// SetClientHandles changes the client handles for one or more items in the collection.
func (is *OPCItems) SetClientHandles(serverHandles []uint32, clientHandles []uint32) []error {
	if is == nil {
		return nil
	}
	resultErrors := make([]error, len(serverHandles))
	for i, handle := range serverHandles {
		item, err := is.GetOPCItem(handle)
		if err != nil {
			resultErrors[i] = err
			continue
		}
		err = item.SetClientHandle(clientHandles[i])
		if err != nil {
			resultErrors[i] = err
		}
	}
	return resultErrors
}

// SetDataTypes changes the requested data type for one or more items in the collection.
func (is *OPCItems) SetDataTypes(serverHandles []uint32, requestedDataTypes []com.VT) []error {
	if is == nil {
		return nil
	}
	resultErrors := make([]error, len(serverHandles))
	for i, handle := range serverHandles {
		item, err := is.GetOPCItem(handle)
		if err != nil {
			resultErrors[i] = err
			continue
		}
		err = item.SetRequestedDataType(requestedDataTypes[i])
		if err != nil {
			resultErrors[i] = err
		}
	}
	return resultErrors
}

// Release releases the OPCItems collection and all associated resources.
func (is *OPCItems) Release() {
	if is == nil {
		return
	}
	for _, item := range is.items {
		item.Release()
	}
	if is.itemMgtProvider != nil {
		is.itemMgtProvider.Release()
	}
}

func (is *OPCItems) createDefinitions(tags []string, accessPath string, active bool, requestedDataType com.VT) []com.TagOPCITEMDEF {
	var definitions []com.TagOPCITEMDEF
	if is == nil {
		return nil
	}
	for _, v := range tags {
		cHandle := atomic.AddUint32(&is.itemID, 1)
		definitions = append(definitions, com.TagOPCITEMDEF{
			SzAccessPath: windows.StringToUTF16Ptr(accessPath),
			SzItemID:     windows.StringToUTF16Ptr(v),
			BActive:      com.BoolToComBOOL(active),
			HClient:      cHandle,
			DwBlobSize:   0,
			PBlob:        nil,
			VtRequested:  uint16(requestedDataType),
		})
	}
	return definitions
}

func (is *OPCItems) getError(errorCode int32) error {
	if is == nil || is.provider == nil {
		return &OPCError{ErrorCode: errorCode, ErrorMessage: "uninitialized common interface"}
	}
	errStr, _ := is.provider.GetErrorString(uint32(errorCode))
	return &OPCError{
		ErrorCode:    errorCode,
		ErrorMessage: errStr,
	}
}
