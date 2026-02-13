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

// OPCServer represents a connection to an OPC DA server.
type OPCServer struct {
	provider   serverProvider
	groups     *OPCGroups
	Name       string     // Name is the ProgID of the server.
	Node       string     // Node is the network node name where the server resides.
	clientName string     // clientName is the name of the client application.
	location   com.CLSCTX // location indicates whether the server is local or remote.

	container *com.IConnectionPointContainer // container manages connection points.
	point     *com.IConnectionPoint          // point is the specific connection point.
	event     *ShutdownEventReceiver         // event receives shutdown notifications.
	cookie    uint32                         // cookie identifies the advisory connection.
}

// Connect establishes a connection to the OPC server.
// It returns an OPCServer instance and an error if connection fails.
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

// newOPCServerWithProvider creates a new OPCServer with a specific provider (used for testing).
func newOPCServerWithProvider(provider serverProvider, name string, node string) *OPCServer {
	s := &OPCServer{
		provider: provider,
		Name:     name,
		Node:     node,
	}
	s.groups = NewOPCGroups(s)
	return s
}

// getClsID retrieves the CLSID from ProgID, trying multiple methods (ServerList V2, V1, Registry).
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

// getClsIDFromServerListV2 attempts to get CLSID using IOPCServerList2.
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

// getClsIDFromServerListV1 attempts to get CLSID using IOPCServerList.
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

// getClsIDFromReg retrieves CLSID directly from Windows Registry.
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

// getClsidFromProgIDKey helper to extract CLSID string and GUID from a registry key.
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

// ServerInfo contains information about an OPC DA server.
type ServerInfo struct {
	ProgID       string        // ProgID is the Program ID of the server.
	ClsStr       string        // ClsStr is the CLSID string representation.
	VerIndProgID string        // VerIndProgID is the Version Independent ProgID.
	ClsID        *windows.GUID // ClsID is the unique Class ID of the server.
}

// GetOPCServers enumerates available OPC servers on a node.
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

// getServersFromOpcServerListV2 enumerates servers using IOPCServerList2.
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

// getServersFromOpcServerListV1 enumerates servers using IOPCServerList.
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

// getServersFromReg enumerates servers by scanning the registry (fallback method).
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

// getServersFromKey helper to extract server info from a registry key.
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

// getServer helper to extract details from IOPCServerList2.
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

// getServerV1 helper to extract details from IOPCServerList.
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

// GetLocaleID returns the current locale ID.
func (s *OPCServer) GetLocaleID() (uint32, error) {
	if s == nil || s.provider == nil {
		return 0, errors.New("uninitialized server connection")
	}
	localeID, err := s.provider.GetLocaleID()
	return localeID, err
}

// GetStartTime returns the time the server started running.
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

// GetCurrentTime returns the current time from the server.
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

// GetLastUpdateTime returns the last update time from the server.
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

// GetMajorVersion returns the major part of the server version number.
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

// GetMinorVersion returns the minor part of the server version number.
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

// GetBuildNumber returns the build number of the server.
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

// GetVendorInfo returns the vendor information string for the server.
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

// GetServerState returns the server's state.
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

// SetLocaleID sets the locale ID.
func (s *OPCServer) SetLocaleID(localeID uint32) error {
	if s == nil || s.provider == nil {
		return errors.New("uninitialized server connection")
	}
	return s.provider.SetLocaleID(localeID)
}

// GetBandwidth returns the bandwidth of the server.
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

// GetOPCGroups returns the collection of OPCGroup objects.
func (s *OPCServer) GetOPCGroups() *OPCGroups {
	if s == nil {
		return nil
	}
	return s.groups
}

// GetServerName returns the server name of the server that the client connected to via Connect().
func (s *OPCServer) GetServerName() string {
	if s == nil {
		return ""
	}
	return s.Name
}

// GetServerNode returns the node name of the server that the client connected to via Connect().
func (s *OPCServer) GetServerNode() string {
	if s == nil {
		return ""
	}
	return s.Node
}

// GetClientName returns the client name of the client.
func (s *OPCServer) GetClientName() string {
	if s == nil {
		return ""
	}
	return s.clientName
}

// SetClientName sets the client name of the client.
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

// PropertyDescription describes an OPC item property.
type PropertyDescription struct {
	PropertyID   int32
	Description  string
	DataType     int16
	AccessRights int16
}

// CreateBrowser creates an OPCBrowser object.
func (s *OPCServer) CreateBrowser() (*OPCBrowser, error) {
	if s == nil || s.provider == nil {
		return nil, errors.New("uninitialized server connection")
	}
	return NewOPCBrowser(s)
}

// GetErrorString converts an error number to a readable string.
func (s *OPCServer) GetErrorString(errorCode int32) (string, error) {
	if s == nil || s.provider == nil {
		return "", errors.New("uninitialized server connection")
	}
	return s.provider.GetErrorString(uint32(errorCode))
}

// QueryAvailableLocaleIDs returns the available LocaleIDs for this server/client session.
func (s *OPCServer) QueryAvailableLocaleIDs() ([]uint32, error) {
	if s == nil || s.provider == nil {
		return nil, errors.New("uninitialized server connection")
	}
	return s.provider.QueryAvailableLocaleIDs()
}

// QueryAvailableProperties returns a list of ID codes and Descriptions for the available properties for this ItemID.
func (s *OPCServer) QueryAvailableProperties(itemID string) (pPropertyIDs []uint32, ppDescriptions []string, ppvtDataTypes []uint16, err error) {
	if s == nil || s.provider == nil {
		return nil, nil, nil, errors.New("uninitialized server connection")
	}
	return s.provider.QueryAvailableProperties(itemID)
}

// GetItemProperties returns a list of the current data values for the passed ID codes.
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

// LookupItemIDs returns a list of ItemIDs (if available) for each of the passed ID codes.
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

// errors converts raw error codes to OPCError structs.
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

// RegisterServerShutDown registers server shut down event.
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

// Disconnect disconnects from the OPC server.
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
