//go:build windows

package opcda

import (
	"unsafe"

	"github.com/wends155/opcda/com"
	"golang.org/x/sys/windows"
)

type serverProvider interface {
	GetStatus() (*com.ServerStatus, error)
	GetErrorString(errorCode uint32) (string, error)
	GetLocaleID() (uint32, error)
	SetLocaleID(localeID uint32) error
	SetClientName(clientName string) error
	QueryAvailableLocaleIDs() ([]uint32, error)
	QueryAvailableProperties(itemID string) ([]uint32, []string, []uint16, error)
	GetItemProperties(itemID string, propertyIDs []uint32) ([]interface{}, []int32, error)
	LookupItemIDs(itemID string, propertyIDs []uint32) ([]string, []int32, error)
	AddGroup(name string, active bool, updateRate uint32, clientGroup uint32, timeBias *int32, deadband *float32, localeID uint32, iid *windows.GUID) (serverGroup uint32, revisedUpdateRate uint32, ppUnk *com.IUnknown, err error)
	RemoveGroup(serverGroup uint32, force bool) error
	Release()
	QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error
}

type comServerProvider struct {
	iServer       *com.IOPCServer
	iCommon       *com.IOPCCommon
	iItemProperty *com.IOPCItemProperties
}

func (p *comServerProvider) GetStatus() (*com.ServerStatus, error) {
	return p.iServer.GetStatus()
}

func (p *comServerProvider) GetErrorString(errorCode uint32) (string, error) {
	return p.iCommon.GetErrorString(errorCode)
}

func (p *comServerProvider) GetLocaleID() (uint32, error) {
	return p.iCommon.GetLocaleID()
}

func (p *comServerProvider) SetLocaleID(localeID uint32) error {
	return p.iCommon.SetLocaleID(localeID)
}

func (p *comServerProvider) SetClientName(clientName string) error {
	return p.iCommon.SetClientName(clientName)
}

func (p *comServerProvider) QueryAvailableLocaleIDs() ([]uint32, error) {
	return p.iCommon.QueryAvailableLocaleIDs()
}

func (p *comServerProvider) QueryAvailableProperties(itemID string) ([]uint32, []string, []uint16, error) {
	return p.iItemProperty.QueryAvailableProperties(itemID)
}

func (p *comServerProvider) GetItemProperties(itemID string, propertyIDs []uint32) ([]interface{}, []int32, error) {
	return p.iItemProperty.GetItemProperties(itemID, propertyIDs)
}

func (p *comServerProvider) LookupItemIDs(itemID string, propertyIDs []uint32) ([]string, []int32, error) {
	return p.iItemProperty.LookupItemIDs(itemID, propertyIDs)
}

func (p *comServerProvider) AddGroup(name string, active bool, updateRate uint32, clientGroup uint32, timeBias *int32, deadband *float32, localeID uint32, iid *windows.GUID) (serverGroup uint32, revisedUpdateRate uint32, ppUnk *com.IUnknown, err error) {
	return p.iServer.AddGroup(name, active, updateRate, clientGroup, timeBias, deadband, localeID, iid)
}

func (p *comServerProvider) RemoveGroup(serverGroup uint32, force bool) error {
	return p.iServer.RemoveGroup(serverGroup, force)
}

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

func (p *comServerProvider) QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error {
	return p.iServer.QueryInterface(iid, ppv)
}
