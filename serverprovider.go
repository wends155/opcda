//go:build windows

package opcda

import (
	"unsafe"

	"github.com/wends155/opcda/com"
	"golang.org/x/sys/windows"
)

// serverProvider defines the internal contract for interacting with the OPC DA server.
// It abstracts the underlying COM implementation to allow for mocking and testing.
type serverProvider interface {
	// GetStatus retrieves the current status of the OPC server.
	GetStatus() (*com.ServerStatus, error)
	// GetErrorString converts an error code to a readable string.
	GetErrorString(errorCode uint32) (string, error)
	// GetLocaleID retrieves the current locale identifier for the server.
	GetLocaleID() (uint32, error)
	// SetLocaleID sets the locale identifier for the server.
	SetLocaleID(localeID uint32) error
	// SetClientName sets the name of the client application.
	SetClientName(clientName string) error
	// QueryAvailableLocaleIDs returns a list of available locale identifiers.
	QueryAvailableLocaleIDs() ([]uint32, error)
	// QueryAvailableProperties returns a list of ID codes and descriptions for the available properties of an item.
	QueryAvailableProperties(itemID string) ([]uint32, []string, []uint16, error)
	// GetItemProperties returns a list of the current data values for the passed ID codes.
	GetItemProperties(itemID string, propertyIDs []uint32) ([]interface{}, []int32, error)
	// LookupItemIDs returns a list of item IDs for each of the passed ID codes.
	LookupItemIDs(itemID string, propertyIDs []uint32) ([]string, []int32, error)
	// AddGroup creates a new OPC group with the specified parameters.
	AddGroup(name string, active bool, updateRate uint32, clientGroup uint32, timeBias *int32, deadband *float32, localeID uint32, iid *windows.GUID) (serverGroup uint32, revisedUpdateRate uint32, ppUnk *com.IUnknown, err error)
	// RemoveGroup removes the specified group from the server.
	RemoveGroup(serverGroup uint32, force bool) error
	// Release releases the COM resources associated with the provider.
	Release()
	// QueryInterface queries the server for a specific interface.
	QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error
}

// comServerProvider is the concrete implementation of serverProvider using COM.
type comServerProvider struct {
	iServer       *com.IOPCServer
	iCommon       *com.IOPCCommon
	iItemProperty *com.IOPCItemProperties
}

// GetStatus retrieves the current status of the OPC server.
func (p *comServerProvider) GetStatus() (*com.ServerStatus, error) {
	return p.iServer.GetStatus()
}

// GetErrorString converts an error code to a readable string.
func (p *comServerProvider) GetErrorString(errorCode uint32) (string, error) {
	return p.iCommon.GetErrorString(errorCode)
}

// GetLocaleID retrieves the current locale identifier for the server.
func (p *comServerProvider) GetLocaleID() (uint32, error) {
	return p.iCommon.GetLocaleID()
}

// SetLocaleID sets the locale identifier for the server.
func (p *comServerProvider) SetLocaleID(localeID uint32) error {
	return p.iCommon.SetLocaleID(localeID)
}

// SetClientName sets the name of the client application.
func (p *comServerProvider) SetClientName(clientName string) error {
	return p.iCommon.SetClientName(clientName)
}

// QueryAvailableLocaleIDs returns a list of available locale identifiers.
func (p *comServerProvider) QueryAvailableLocaleIDs() ([]uint32, error) {
	return p.iCommon.QueryAvailableLocaleIDs()
}

// QueryAvailableProperties returns a list of ID codes and descriptions for the available properties of an item.
func (p *comServerProvider) QueryAvailableProperties(itemID string) ([]uint32, []string, []uint16, error) {
	return p.iItemProperty.QueryAvailableProperties(itemID)
}

// GetItemProperties returns a list of the current data values for the passed ID codes.
func (p *comServerProvider) GetItemProperties(itemID string, propertyIDs []uint32) ([]interface{}, []int32, error) {
	return p.iItemProperty.GetItemProperties(itemID, propertyIDs)
}

// LookupItemIDs returns a list of item IDs for each of the passed ID codes.
func (p *comServerProvider) LookupItemIDs(itemID string, propertyIDs []uint32) ([]string, []int32, error) {
	return p.iItemProperty.LookupItemIDs(itemID, propertyIDs)
}

// AddGroup creates a new OPC group with the specified parameters.
func (p *comServerProvider) AddGroup(name string, active bool, updateRate uint32, clientGroup uint32, timeBias *int32, deadband *float32, localeID uint32, iid *windows.GUID) (serverGroup uint32, revisedUpdateRate uint32, ppUnk *com.IUnknown, err error) {
	return p.iServer.AddGroup(name, active, updateRate, clientGroup, timeBias, deadband, localeID, iid)
}

// RemoveGroup removes the specified group from the server.
func (p *comServerProvider) RemoveGroup(serverGroup uint32, force bool) error {
	return p.iServer.RemoveGroup(serverGroup, force)
}

// Release releases the COM resources associated with the provider.
func (p *comServerProvider) Release() {
	if p.iItemProperty != nil {
		p.iItemProperty.Release()
	}
	if p.iCommon != nil {
		p.iCommon.Release()
	}
	if p.iServer != nil {
		p.iServer.Release()
	}
}

// QueryInterface queries the server for a specific interface.
func (p *comServerProvider) QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error {
	return p.iServer.QueryInterface(iid, ppv)
}
