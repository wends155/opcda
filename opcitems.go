//go:build windows

package opcda

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/wends155/opcda/com"

	"golang.org/x/sys/windows"
)

type itemMgtProvider interface {
	AddItems(items []com.TagOPCITEMDEF) ([]com.TagOPCITEMRESULTStruct, []int32, error)
	ValidateItems(items []com.TagOPCITEMDEF, bBlob bool) ([]com.TagOPCITEMRESULTStruct, []int32, error)
	RemoveItems(serverHandles []uint32) ([]int32, error)
	SetActiveState(serverHandles []uint32, bActive bool) ([]int32, error)
	SetClientHandles(serverHandles []uint32, clientHandles []uint32) ([]int32, error)
	SetDatatypes(serverHandles []uint32, requestedDataTypes []com.VT) ([]int32, error)
	Release()
}

type comItemMgtProvider struct {
	itemMgt *com.IOPCItemMgt
}

func (p *comItemMgtProvider) AddItems(items []com.TagOPCITEMDEF) ([]com.TagOPCITEMRESULTStruct, []int32, error) {
	return p.itemMgt.AddItems(items)
}

func (p *comItemMgtProvider) ValidateItems(items []com.TagOPCITEMDEF, bBlob bool) ([]com.TagOPCITEMRESULTStruct, []int32, error) {
	return p.itemMgt.ValidateItems(items, bBlob)
}

func (p *comItemMgtProvider) RemoveItems(serverHandles []uint32) ([]int32, error) {
	return p.itemMgt.RemoveItems(serverHandles)
}

func (p *comItemMgtProvider) SetActiveState(serverHandles []uint32, bActive bool) ([]int32, error) {
	return p.itemMgt.SetActiveState(serverHandles, bActive)
}

func (p *comItemMgtProvider) SetClientHandles(serverHandles []uint32, clientHandles []uint32) ([]int32, error) {
	return p.itemMgt.SetClientHandles(serverHandles, clientHandles)
}

func (p *comItemMgtProvider) SetDatatypes(serverHandles []uint32, requestedDataTypes []com.VT) ([]int32, error) {
	return p.itemMgt.SetDatatypes(serverHandles, requestedDataTypes)
}

func (p *comItemMgtProvider) Release() {
	p.itemMgt.Release()
}

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

// GetParent Returns reference to the parent OPCGroup object
func (is *OPCItems) GetParent() *OPCGroup {
	if is == nil {
		return nil
	}
	return is.parent
}

// GetDefaultRequestedDataType get the requested data type that will be used in calls to Add
func (is *OPCItems) GetDefaultRequestedDataType() com.VT {
	if is == nil {
		return com.VT_EMPTY
	}
	return is.defaultRequestedDataType
}

// SetDefaultRequestedDataType set the requested data type that will be used in calls to Add
func (is *OPCItems) SetDefaultRequestedDataType(defaultRequestedDataType com.VT) {
	if is == nil {
		return
	}
	is.defaultRequestedDataType = defaultRequestedDataType
}

// GetDefaultAccessPath get the default AccessPath that will be used in calls to Add
func (is *OPCItems) GetDefaultAccessPath() string {
	if is == nil {
		return ""
	}
	return is.defaultAccessPath
}

// SetDefaultAccessPath set the default AccessPath that will be used in calls to Add
func (is *OPCItems) SetDefaultAccessPath(defaultAccessPath string) {
	if is == nil {
		return
	}
	is.Lock()
	defer is.Unlock()
	is.defaultAccessPath = defaultAccessPath
}

// GetDefaultActive get the default active state for OPCItems created using Items.Add
func (is *OPCItems) GetDefaultActive() bool {
	if is == nil {
		return false
	}
	return is.defaultActive
}

// SetDefaultActive set the default active state for OPCItems created using Items.Add
func (is *OPCItems) SetDefaultActive(defaultActive bool) {
	if is == nil {
		return
	}
	is.defaultActive = defaultActive
}

// GetCount get the number of items in the collection
func (is *OPCItems) GetCount() int {
	if is == nil {
		return 0
	}
	is.RLock()
	defer is.RUnlock()
	return len(is.items)
}

// Item get the item by index
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

// ItemByName get the item by name
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

// GetOPCItem returns the OPCItem by serverHandle
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

// AddItems adds items to the group.
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

// Remove Removes an OPCItem
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

// Validate Determines if one or more OPCItems could be successfully created via the Add method (but does not add them).
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

// SetActive Allows Activation and deactivation of individual OPCItem’s in the OPCItems Collection
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

// SetClientHandles Changes the client handles or one or more Items in a Group.
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

// SetDataTypes Changes the requested data type for one or more Items
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

// Release Releases the OPCItems collection and all associated resources.
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
