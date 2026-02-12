//go:build windows

package opcda

import (
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/wends155/opcda/com"
	"golang.org/x/sys/windows/registry"

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

type OPCServer struct {
	provider   serverProvider
	groups     *OPCGroups
	Name       string
	Node       string
	clientName string
	location   com.CLSCTX

	container *com.IConnectionPointContainer
	point     *com.IConnectionPoint
	event     *ShutdownEventReceiver
	cookie    uint32
}

// Connect connect to OPC server
func Connect(progID, node string) (opcServer *OPCServer, err error) {
	location := com.CLSCTX_LOCAL_SERVER
	if !com.IsLocal(node) {
		location = com.CLSCTX_REMOTE_SERVER
	}
	clsid, err := getClsID(progID, node, location)
	if err != nil {
		return nil, NewOPCWrapperError("get clsid", err)
	}
	iUnknownServer, err := com.MakeCOMObjectEx(node, location, clsid, &com.IID_IOPCServer)
	if err != nil {
		return nil, NewOPCWrapperError("make com object IOPCServer", err)
	}
	defer func() {
		if err != nil {
			iUnknownServer.Release()
		}
	}()
	var iUnknownCommon *com.IUnknown
	err = iUnknownServer.QueryInterface(&com.IID_IOPCCommon, unsafe.Pointer(&iUnknownCommon))
	if err != nil {
		return nil, NewOPCWrapperError("server query interface IOPCCommon", err)
	}
	defer func() {
		if err != nil {
			iUnknownCommon.Release()
		}
	}()
	var iUnknownItemProperties *com.IUnknown
	err = iUnknownServer.QueryInterface(&com.IID_IOPCItemProperties, unsafe.Pointer(&iUnknownItemProperties))
	if err != nil {
		return nil, NewOPCWrapperError("server query interface IOPCItemProperties", err)
	}
	defer func() {
		if err != nil {
			iUnknownItemProperties.Release()
		}
	}()
	server := &com.IOPCServer{IUnknown: iUnknownServer}
	common := &com.IOPCCommon{IUnknown: iUnknownCommon}
	itemProperties := &com.IOPCItemProperties{IUnknown: iUnknownItemProperties}
	opcServer = &OPCServer{
		provider: &comServerProvider{
			iServer:       server,
			iCommon:       common,
			iItemProperty: itemProperties,
		},
		Name:     progID,
		Node:     node,
		location: location,
	}
	opcServer.groups = NewOPCGroups(opcServer)
	return opcServer, nil
}

func newOPCServerWithProvider(provider serverProvider, name string, node string) *OPCServer {
	s := &OPCServer{
		provider: provider,
		Name:     name,
		Node:     node,
	}
	s.groups = NewOPCGroups(s)
	return s
}

func getClsID(progID, node string, location com.CLSCTX) (clsid *windows.GUID, err error) {
	var errorList []error
	// try get clsid from server list
	clsid, err = getClsIDFromServerListV2(progID, node, location)
	if err == nil {
		return clsid, nil
	}
	errorList = append(errorList, fmt.Errorf("get clsid from server list v2 error: %v", err))
	// try v1
	clsid, err = getClsIDFromServerListV1(progID, node, location)
	if err == nil {
		return clsid, nil
	}
	errorList = append(errorList, fmt.Errorf("get clsid from server list v1 error: %v", err))
	// try get clsid from windows reg
	clsid, err = getClsIDFromReg(progID, node)
	if err == nil {
		return clsid, nil
	}
	errorList = append(errorList, fmt.Errorf("get clsid from reg error: %v", err))
	return nil, errors.Join(errorList...)
}

func getClsIDFromServerListV2(progID, node string, location com.CLSCTX) (*windows.GUID, error) {
	iCatInfo, err := com.MakeCOMObjectEx(node, location, &com.CLSID_OpcServerList, &com.IID_IOPCServerList2)
	if err != nil {
		return nil, err
	}
	defer iCatInfo.Release()
	sl := &com.IOPCServerList2{IUnknown: iCatInfo}
	clsid, err := sl.CLSIDFromProgID(progID)
	if err != nil {
		return nil, err
	}
	return clsid, nil
}

func getClsIDFromServerListV1(progID, node string, location com.CLSCTX) (*windows.GUID, error) {
	iCatInfo, err := com.MakeCOMObjectEx(node, location, &com.CLSID_OpcServerList, &com.IID_IOPCServerList)
	if err != nil {
		return nil, err
	}
	defer iCatInfo.Release()
	sl := &com.IOPCServerList{IUnknown: iCatInfo}
	clsid, err := sl.CLSIDFromProgID(progID)
	if err != nil {
		return nil, err
	}
	return clsid, nil
}

func getClsIDFromReg(progID, node string) (*windows.GUID, error) {
	var clsid *windows.GUID
	var err error
	hKey, err := registry.OpenRemoteKey(node, registry.CLASSES_ROOT)
	if err != nil {
		return nil, err
	}
	defer hKey.Close()

	hProgIDKey, err := registry.OpenKey(hKey, progID, registry.READ)
	if err != nil {
		return nil, err
	}
	defer hProgIDKey.Close()
	_, clsid, err = getClsidFromProgIDKey(hProgIDKey)
	return clsid, err
}

func getClsidFromProgIDKey(hProgIDKey registry.Key) (string, *windows.GUID, error) {
	hClsidKey, err := registry.OpenKey(hProgIDKey, "CLSID", registry.READ)
	if err != nil {
		return "", nil, err
	}
	defer hClsidKey.Close()
	clsidStr, _, err := hClsidKey.GetStringValue("")
	if err != nil {
		return "", nil, err
	}
	clsid, err := windows.GUIDFromString(clsidStr)
	return clsidStr, &clsid, err
}

type ServerInfo struct {
	ProgID       string
	ClsStr       string
	VerIndProgID string
	ClsID        *windows.GUID
}

// GetOPCServers get OPC servers from node
func GetOPCServers(node string) ([]*ServerInfo, error) {
	var errorList []error
	result, err := getServersFromOpcServerListV2(node)
	if err == nil {
		return result, nil
	}
	errorList = append(errorList, fmt.Errorf("get servers from opc server list v2 error: %v", err))
	// try v1
	result, err = getServersFromOpcServerListV1(node)
	if err == nil {
		return result, nil
	}
	errorList = append(errorList, fmt.Errorf("get servers from opc server list v1 error: %v", err))
	// try windows reg
	result, err = getServersFromReg(node)
	if err == nil {
		return result, nil
	}
	errorList = append(errorList, fmt.Errorf("get servers from reg error: %v", err))
	return nil, errors.Join(errorList...)
}

func getServersFromOpcServerListV2(node string) ([]*ServerInfo, error) {
	location := com.CLSCTX_LOCAL_SERVER
	if !com.IsLocal(node) {
		location = com.CLSCTX_REMOTE_SERVER
	}
	iCatInfo, err := com.MakeCOMObjectEx(node, location, &com.CLSID_OpcServerList, &com.IID_IOPCServerList2)
	if err != nil {
		return nil, NewOPCWrapperError("make com object IOPCServerListV2", err)
	}
	cids := []windows.GUID{IID_CATID_OPCDAServer10, IID_CATID_OPCDAServer20}
	defer iCatInfo.Release()
	sl := &com.IOPCServerList2{IUnknown: iCatInfo}
	iEnum, err := sl.EnumClassesOfCategories(cids, nil)
	if err != nil {
		return nil, NewOPCWrapperError("enum classes of categories with IOPCServerListV2", err)
	}
	defer iEnum.Release()
	var result []*ServerInfo
	for {
		var classID windows.GUID
		var actual uint32
		err = iEnum.Next(1, &classID, &actual)
		if err != nil {
			break
		}
		server, err := getServer(sl, &classID)
		if err != nil {
			return nil, NewOPCWrapperError("IOPCServerListV2 getServer", err)
		}
		result = append(result, server)
	}
	return result, nil
}

func getServersFromOpcServerListV1(node string) ([]*ServerInfo, error) {
	location := com.CLSCTX_LOCAL_SERVER
	if !com.IsLocal(node) {
		location = com.CLSCTX_REMOTE_SERVER
	}
	iCatInfo, err := com.MakeCOMObjectEx(node, location, &com.CLSID_OpcServerList, &com.IID_IOPCServerList)
	if err != nil {
		return nil, NewOPCWrapperError("make com object IOPCServerListV1", err)
	}
	cids := []windows.GUID{IID_CATID_OPCDAServer10, IID_CATID_OPCDAServer20}
	defer iCatInfo.Release()
	sl := &com.IOPCServerList{IUnknown: iCatInfo}
	iEnum, err := sl.EnumClassesOfCategories(cids, nil)
	if err != nil {
		return nil, NewOPCWrapperError("enum classes of categories with IOPCServerListV1", err)
	}
	defer iEnum.Release()
	var result []*ServerInfo
	for {
		var classID windows.GUID
		var actual uint32
		err = iEnum.Next(1, &classID, &actual)
		if err != nil {
			break
		}
		server, err := getServerV1(sl, &classID)
		if err != nil {
			return nil, NewOPCWrapperError("IOPCServerListV1 getServer", err)
		}
		result = append(result, server)
	}
	return result, nil
}

func getServersFromReg(node string) ([]*ServerInfo, error) {
	var result []*ServerInfo
	var hKey registry.Key
	var err error
	if node == "" || node == "localhost" {
		hKey = registry.CLASSES_ROOT
	} else {
		hKey, err = registry.OpenRemoteKey(node, registry.CLASSES_ROOT)
		if err != nil {
			return nil, err
		}
		defer hKey.Close()
	}
	tsKeys, _ := hKey.ReadSubKeyNames(-1)
	for _, tsKey := range tsKeys {
		info := getServersFromKey(hKey, tsKey)
		if info != nil {
			result = append(result, info)
		}
	}
	return result, nil
}

func getServersFromKey(hKey registry.Key, progID string) *ServerInfo {
	hProgIDKey, err := registry.OpenKey(hKey, progID, registry.READ)
	if err != nil {
		return nil
	}
	defer hProgIDKey.Close()
	hOPCKey, err := registry.OpenKey(hProgIDKey, "OPC", registry.READ)
	if err != nil {
		return nil
	}
	defer hOPCKey.Close()
	clsidStr, clsid, err := getClsidFromProgIDKey(hProgIDKey)
	if err != nil {
		return nil
	}
	return &ServerInfo{
		ProgID:       progID,
		ClsStr:       clsidStr,
		VerIndProgID: progID,
		ClsID:        clsid,
	}
}

func getServer(sl *com.IOPCServerList2, classID *windows.GUID) (*ServerInfo, error) {
	progID, userType, VerIndProgID, err := sl.GetClassDetails(classID)
	if err != nil {
		return nil, fmt.Errorf("FAILED to get prog ID from class ID: %w", err)
	}
	defer func() {
		com.CoTaskMemFree(unsafe.Pointer(progID))
		com.CoTaskMemFree(unsafe.Pointer(userType))
		com.CoTaskMemFree(unsafe.Pointer(VerIndProgID))
	}()
	clsStr := classID.String()
	return &ServerInfo{
		ProgID:       windows.UTF16PtrToString(progID),
		ClsStr:       clsStr,
		ClsID:        classID,
		VerIndProgID: windows.UTF16PtrToString(VerIndProgID),
	}, nil
}

func getServerV1(sl *com.IOPCServerList, classID *windows.GUID) (*ServerInfo, error) {
	progID, userType, err := sl.GetClassDetails(classID)
	if err != nil {
		return nil, fmt.Errorf("FAILED to get prog ID from class ID: %w", err)
	}
	defer func() {
		com.CoTaskMemFree(unsafe.Pointer(progID))
		com.CoTaskMemFree(unsafe.Pointer(userType))
	}()
	clsStr := classID.String()
	return &ServerInfo{
		ProgID:       windows.UTF16PtrToString(progID),
		ClsStr:       clsStr,
		ClsID:        classID,
		VerIndProgID: "",
	}, nil
}

// GetLocaleID get locale ID
func (s *OPCServer) GetLocaleID() (uint32, error) {
	if s == nil || s.provider == nil {
		return 0, errors.New("uninitialized server connection")
	}
	localeID, err := s.provider.GetLocaleID()
	return localeID, err
}

// GetStartTime Returns the time the server started running
func (s *OPCServer) GetStartTime() (time.Time, error) {
	if s == nil || s.provider == nil {
		return time.Time{}, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return time.Time{}, err
	}
	return status.StartTime, nil
}

// GetCurrentTime Returns the current time from the server
func (s *OPCServer) GetCurrentTime() (time.Time, error) {
	if s == nil || s.provider == nil {
		return time.Time{}, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return time.Time{}, err
	}
	return status.CurrentTime, nil
}

// GetLastUpdateTime Returns the last update time from the server
func (s *OPCServer) GetLastUpdateTime() (time.Time, error) {
	if s == nil || s.provider == nil {
		return time.Time{}, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return time.Time{}, err
	}
	return status.LastUpdateTime, nil
}

// GetMajorVersion Returns the major part of the server version number
func (s *OPCServer) GetMajorVersion() (uint16, error) {
	if s == nil || s.provider == nil {
		return 0, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return 0, err
	}
	return status.MajorVersion, nil
}

// GetMinorVersion Returns the minor part of the server version number
func (s *OPCServer) GetMinorVersion() (uint16, error) {
	if s == nil || s.provider == nil {
		return 0, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return 0, err
	}
	return status.MinorVersion, nil
}

// GetBuildNumber Returns the build number of the server
func (s *OPCServer) GetBuildNumber() (uint16, error) {
	if s == nil || s.provider == nil {
		return 0, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return 0, err
	}
	return status.BuildNumber, nil
}

// GetVendorInfo Returns the vendor information string for the server
func (s *OPCServer) GetVendorInfo() (string, error) {
	if s == nil || s.provider == nil {
		return "", errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return "", err
	}
	return status.VendorInfo, nil
}

// GetServerState Returns the server’s state
func (s *OPCServer) GetServerState() (com.OPCServerState, error) {
	if s == nil || s.provider == nil {
		return 0, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return 0, err
	}
	return status.ServerState, nil
}

// SetLocaleID set locale ID
func (s *OPCServer) SetLocaleID(localeID uint32) error {
	if s == nil || s.provider == nil {
		return errors.New("uninitialized server connection")
	}
	return s.provider.SetLocaleID(localeID)
}

// GetBandwidth Returns the bandwidth of the server
func (s *OPCServer) GetBandwidth() (uint32, error) {
	if s == nil || s.provider == nil {
		return 0, errors.New("uninitialized server connection")
	}
	status, err := s.provider.GetStatus()
	if err != nil {
		return 0, err
	}
	return status.BandWidth, nil
}

// GetOPCGroups get a collection of OPCGroup objects
func (s *OPCServer) GetOPCGroups() *OPCGroups {
	if s == nil {
		return nil
	}
	return s.groups
}

// GetServerName Returns the server name of the server that the client connected to via Connect().
func (s *OPCServer) GetServerName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

// GetServerNode Returns the node name of the server that the client connected to via Connect().
func (s *OPCServer) GetServerNode() string {
	if s == nil {
		return ""
	}
	return s.Node
}

// GetClientName Returns the client name of the client
func (s *OPCServer) GetClientName() string {
	if s == nil {
		return ""
	}
	return s.clientName
}

// SetClientName Sets the client name of the client
func (s *OPCServer) SetClientName(clientName string) error {
	if s == nil || s.provider == nil {
		return errors.New("uninitialized server connection")
	}
	err := s.provider.SetClientName(clientName)
	if err != nil {
		return err
	}
	s.clientName = clientName
	return nil
}

type PropertyDescription struct {
	PropertyID   int32
	Description  string
	DataType     int16
	AccessRights int16
}

// CreateBrowser Creates an OPCBrowser object
func (s *OPCServer) CreateBrowser() (*OPCBrowser, error) {
	if s == nil || s.provider == nil {
		return nil, errors.New("uninitialized server connection")
	}
	return NewOPCBrowser(s)
}

// GetErrorString Converts an error number to a readable string
func (s *OPCServer) GetErrorString(errorCode int32) (string, error) {
	if s == nil || s.provider == nil {
		return "", errors.New("uninitialized server connection")
	}
	return s.provider.GetErrorString(uint32(errorCode))
}

// QueryAvailableLocaleIDs Return the available LocaleIDs for this server/client session
func (s *OPCServer) QueryAvailableLocaleIDs() ([]uint32, error) {
	if s == nil || s.provider == nil {
		return nil, errors.New("uninitialized server connection")
	}
	return s.provider.QueryAvailableLocaleIDs()
}

// QueryAvailableProperties Return a list of ID codes and Descriptions for the available properties for this ItemID
func (s *OPCServer) QueryAvailableProperties(itemID string) (pPropertyIDs []uint32, ppDescriptions []string, ppvtDataTypes []uint16, err error) {
	if s == nil || s.provider == nil {
		return nil, nil, nil, errors.New("uninitialized server connection")
	}
	return s.provider.QueryAvailableProperties(itemID)
}

// GetItemProperties Return a list of the current data values for the passed ID codes.
func (s *OPCServer) GetItemProperties(itemID string, propertyIDs []uint32) (data []interface{}, itemErrors []error, err error) {
	if s == nil || s.provider == nil {
		return nil, nil, errors.New("uninitialized server connection")
	}
	var errs []int32
	data, errs, err = s.provider.GetItemProperties(itemID, propertyIDs)
	if err != nil {
		return nil, nil, err
	}
	itemErrors = s.errors(errs)
	return data, itemErrors, nil
}

// LookupItemIDs Return a list of ItemIDs (if available) for each of the passed ID codes.
// have not tested because simulator return error
func (s *OPCServer) LookupItemIDs(itemID string, propertyIDs []uint32) ([]string, []error, error) {
	if s == nil || s.provider == nil {
		return nil, nil, errors.New("uninitialized server connection")
	}
	ItemIDs, errs, err := s.provider.LookupItemIDs(itemID, propertyIDs)
	if err != nil {
		return nil, nil, err
	}
	itemErrors := s.errors(errs)
	return ItemIDs, itemErrors, nil
}

func (s *OPCServer) errors(errs []int32) []error {
	errors := make([]error, len(errs))
	for i, e := range errs {
		if e < 0 {
			errStr, _ := s.GetErrorString(e)
			errors[i] = &OPCError{
				ErrorCode:    e,
				ErrorMessage: errStr,
			}
		}
	}
	return errors
}

// RegisterServerShutDown register server shut down event
func (s *OPCServer) RegisterServerShutDown(ch chan string) error {
	if s == nil || s.provider == nil {
		return errors.New("uninitialized server connection")
	}
	if s.event == nil {
		var err error
		var iUnknownContainer *com.IUnknown
		var point *com.IConnectionPoint
		var cookie uint32

		err = s.provider.QueryInterface(&com.IID_IConnectionPointContainer, unsafe.Pointer(&iUnknownContainer))
		if err != nil {
			return NewOPCWrapperError("query interface IConnectionPointContainer", err)
		}
		defer func() {
			if err != nil {
				iUnknownContainer.Release()
			}
		}()
		container := &com.IConnectionPointContainer{IUnknown: iUnknownContainer}
		point, err = container.FindConnectionPoint(&IID_IOPCShutdown)
		if err != nil {
			return NewOPCWrapperError("container find connect point", err)
		}
		defer func() {
			if err != nil {
				point.Release()
			}
		}()
		event := NewShutdownEventReceiver()
		cookie, err = point.Advise((*com.IUnknown)(unsafe.Pointer(event)))
		if err != nil {
			return NewOPCWrapperError("point advise", err)
		}
		s.container = container
		s.point = point
		s.event = event
		s.cookie = cookie
	}
	s.event.AddReceiver(ch)
	return nil
}

// Disconnect from OPC server
func (s *OPCServer) Disconnect() error {
	if s == nil {
		return nil
	}
	var err error
	if s.point != nil {
		err = s.point.Unadvise(s.cookie)
		s.point.Release()
	}
	if s.container != nil {
		s.container.Release()
	}
	if s.groups != nil {
		s.groups.Release()
	}
	if s.provider != nil {
		s.provider.Release()
	}
	return err
}
