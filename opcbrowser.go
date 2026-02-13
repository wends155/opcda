//go:build windows

package opcda

import (
	"errors"
	"unsafe"

	"github.com/wends155/opcda/com"
)

// browserProvider defines the methods required for browsing OPC item IDs.
// It abstracts the underlying COM implementation (com.IOPCBrowseServerAddressSpace)
// to enable unit testing and mocking.
//
// Example usage in tests:
//
//	mock := &mockBrowserProvider{}
//	browser := &OPCBrowser{provider: mock}
type browserProvider interface {
	GetItemID(szItemDataID string) (string, error)
	QueryOrganization() (com.OPCNAMESPACETYPE, error)
	BrowseOPCItemIDs(dwBrowseFilterType com.OPCBROWSETYPE, szFilterCriteria string, vtDataTypeFilter uint16, dwAccessRightsFilter uint32) ([]string, error)
	ChangeBrowsePosition(dwBrowseDirection com.OPCBROWSEDIRECTION, szString string) error
	Release()
}

type comBrowserProvider struct {
	iBrowseServerAddressSpace *com.IOPCBrowseServerAddressSpace
}

func (p *comBrowserProvider) GetItemID(szItemDataID string) (string, error) {
	return p.iBrowseServerAddressSpace.GetItemID(szItemDataID)
}

func (p *comBrowserProvider) QueryOrganization() (com.OPCNAMESPACETYPE, error) {
	return p.iBrowseServerAddressSpace.QueryOrganization()
}

func (p *comBrowserProvider) BrowseOPCItemIDs(dwBrowseFilterType com.OPCBROWSETYPE, szFilterCriteria string, vtDataTypeFilter uint16, dwAccessRightsFilter uint32) ([]string, error) {
	return p.iBrowseServerAddressSpace.BrowseOPCItemIDs(dwBrowseFilterType, szFilterCriteria, vtDataTypeFilter, dwAccessRightsFilter)
}

func (p *comBrowserProvider) ChangeBrowsePosition(dwBrowseDirection com.OPCBROWSEDIRECTION, szString string) error {
	return p.iBrowseServerAddressSpace.ChangeBrowsePosition(dwBrowseDirection, szString)
}

func (p *comBrowserProvider) Release() {
	if p.iBrowseServerAddressSpace != nil {
		p.iBrowseServerAddressSpace.Release()
	}
}

type OPCBrowser struct {
	provider     browserProvider
	filter       string
	dataType     uint16
	accessRights uint32
	names        []string
	parent       *OPCServer
}

func NewOPCBrowser(parent *OPCServer) (*OPCBrowser, error) {
	if parent == nil || parent.provider == nil {
		return nil, errors.New("parent server is nil or uninitialized")
	}
	var iBrowseServerAddressSpace *com.IUnknown
	err := parent.provider.QueryInterface(&com.IID_IOPCBrowseServerAddressSpace, unsafe.Pointer(&iBrowseServerAddressSpace))
	if err != nil {
		return nil, NewOPCWrapperError("query interface IOPCBrowseServerAddressSpace", err)
	}
	return newOPCBrowserWithProvider(&comBrowserProvider{iBrowseServerAddressSpace: &com.IOPCBrowseServerAddressSpace{IUnknown: iBrowseServerAddressSpace}}, parent), nil
}

func newOPCBrowserWithProvider(provider browserProvider, parent *OPCServer) *OPCBrowser {
	return &OPCBrowser{
		provider:     provider,
		parent:       parent,
		accessRights: OPC_READABLE | OPC_WRITEABLE,
	}
}

// GetFilter get the filter that applies to ShowBranches and ShowLeafs methods
func (b *OPCBrowser) GetFilter() string {
	if b == nil {
		return ""
	}
	return b.filter
}

// SetFilter set the filter that applies to ShowBranches and ShowLeafs methods
func (b *OPCBrowser) SetFilter(filter string) {
	if b == nil {
		return
	}
	b.filter = filter
}

// GetDataType get the requested data type that applies to ShowLeafs methods. This property defaults to
// com.VT_EMPTY, which means that any data type is acceptable.
func (b *OPCBrowser) GetDataType() uint16 {
	if b == nil {
		return 0
	}
	return b.dataType
}

// SetDataType set the requested data type that applies to ShowLeafs methods.
func (b *OPCBrowser) SetDataType(dataType uint16) {
	if b == nil {
		return
	}
	b.dataType = dataType
}

// GetAccessRights get the requested access rights that apply to the ShowLeafs methods
func (b *OPCBrowser) GetAccessRights() uint32 {
	if b == nil {
		return 0
	}
	return b.accessRights
}

// SetAccessRights set the requested access rights that apply to the ShowLeafs methods
func (b *OPCBrowser) SetAccessRights(accessRights uint32) error {
	if b == nil {
		return errors.New("uninitialized browser")
	}
	if accessRights&OPC_READABLE == 0 && accessRights&OPC_WRITEABLE == 0 {
		return errors.New("accessRights must be OPC_READABLE or OPC_WRITEABLE")
	}
	b.accessRights = accessRights
	return nil
}

// GetCurrentPosition Returns the current position in the tree
func (b *OPCBrowser) GetCurrentPosition() (string, error) {
	if b == nil || b.provider == nil {
		return "", errors.New("uninitialized browser")
	}
	id, err := b.provider.GetItemID("")
	return id, err
}

// GetOrganization Returns either OPCHierarchical or OPCFlat.
func (b *OPCBrowser) GetOrganization() (com.OPCNAMESPACETYPE, error) {
	if b == nil || b.provider == nil {
		return 0, errors.New("uninitialized browser")
	}
	return b.provider.QueryOrganization()
}

// GetCount Required property for collections
func (b *OPCBrowser) GetCount() int {
	if b == nil {
		return 0
	}
	return len(b.names)
}

// Item returns the name of the item at the specified index. index is 0-based.
func (b *OPCBrowser) Item(index int) (string, error) {
	if b == nil {
		return "", errors.New("uninitialized browser")
	}
	if index < 0 || index >= len(b.names) {
		return "", errors.New("index out of range")
	}
	return b.names[index], nil
}

// ShowBranches Fills the collection with names of the branches at the current browse position.
func (b *OPCBrowser) ShowBranches() error {
	if b == nil || b.provider == nil {
		return errors.New("uninitialized browser")
	}
	b.names = nil
	var err error
	b.names, err = b.provider.BrowseOPCItemIDs(OPC_BRANCH, b.filter, b.dataType, b.accessRights)
	return err
}

// ShowLeafs Fills the collection with the names of the leafs at the current browse position
func (b *OPCBrowser) ShowLeafs(flat bool) error {
	if b == nil || b.provider == nil {
		return errors.New("uninitialized browser")
	}
	b.names = nil
	var err error
	browseType := OPC_LEAF
	if flat {
		browseType = OPC_FLAT
	}
	b.names, err = b.provider.BrowseOPCItemIDs(browseType, b.filter, b.dataType, b.accessRights)
	return err
}

// MoveUp Move up one level in the tree.
func (b *OPCBrowser) MoveUp() error {
	if b == nil || b.provider == nil {
		return errors.New("uninitialized browser")
	}
	return b.provider.ChangeBrowsePosition(OPC_BROWSE_UP, "")
}

// MoveToRoot Move up to the first level in the tree.
func (b *OPCBrowser) MoveToRoot() {
	if b == nil || b.provider == nil {
		return
	}
	for {
		err := b.provider.ChangeBrowsePosition(OPC_BROWSE_UP, "")
		if err != nil {
			break
		}
	}
}

// MoveDown Move down into this branch.
func (b *OPCBrowser) MoveDown(name string) error {
	if b == nil || b.provider == nil {
		return errors.New("uninitialized browser")
	}
	return b.provider.ChangeBrowsePosition(OPC_BROWSE_DOWN, name)
}

// MoveTo Move to an absolute position.
func (b *OPCBrowser) MoveTo(branches []string) error {
	if b == nil || b.provider == nil {
		return errors.New("uninitialized browser")
	}
	b.MoveToRoot()
	for _, branch := range branches {
		err := b.MoveDown(branch)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetItemID Given a name, returns a valid ItemID that can be passed to OPCItems Add method.
func (b *OPCBrowser) GetItemID(leaf string) (string, error) {
	if b == nil || b.provider == nil {
		return "", errors.New("uninitialized browser")
	}
	return b.provider.GetItemID(leaf)
}

// Release release the OPCBrowser
func (b *OPCBrowser) Release() {
	if b == nil || b.provider == nil {
		return
	}
	b.provider.Release()
}
