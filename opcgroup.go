//go:build windows

package opcda

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/wends155/opcda/com"
	"golang.org/x/sys/windows"
)

type groupProvider interface {
	SetName(name string) error
	GetState() (updateRate uint32, active bool, name string, timeBias int32, deadband float32, localeID uint32, clientHandle uint32, serverHandle uint32, err error)
	SetState(pRequestedUpdateRate *uint32, pActive *int32, pTimeBias *int32, pPercentDeadband *float32, pLCID *uint32, phClientGroup *uint32) (pRevisedUpdateRate uint32, err error)
	SyncRead(source com.OPCDATASOURCE, serverHandles []uint32) ([]*com.ItemState, []int32, error)
	SyncWrite(serverHandles []uint32, values []com.VARIANT) ([]int32, error)
	AsyncRead(serverHandles []uint32, transactionID uint32) (cancelID uint32, errs []int32, err error)
	AsyncWrite(serverHandles []uint32, values []com.VARIANT, transactionID uint32) (cancelID uint32, errs []int32, err error)
	AsyncRefresh(source com.OPCDATASOURCE, transactionID uint32) (cancelID uint32, err error)
	AsyncCancel(cancelID uint32) error
	QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error
	Release()
}

type comGroupProvider struct {
	groupStateMgt *com.IOPCGroupStateMgt
	syncIO        *com.IOPCSyncIO
	asyncIO2      *com.IOPCAsyncIO2
}

func (p *comGroupProvider) SetName(name string) error {
	return p.groupStateMgt.SetName(name)
}

func (p *comGroupProvider) GetState() (uint32, bool, string, int32, float32, uint32, uint32, uint32, error) {
	return p.groupStateMgt.GetState()
}

func (p *comGroupProvider) SetState(pRequestedUpdateRate *uint32, pActive *int32, pTimeBias *int32, pPercentDeadband *float32, pLCID *uint32, phClientGroup *uint32) (uint32, error) {
	return p.groupStateMgt.SetState(pRequestedUpdateRate, pActive, pTimeBias, pPercentDeadband, pLCID, phClientGroup)
}

func (p *comGroupProvider) SyncRead(source com.OPCDATASOURCE, serverHandles []uint32) ([]*com.ItemState, []int32, error) {
	return p.syncIO.Read(source, serverHandles)
}

func (p *comGroupProvider) SyncWrite(serverHandles []uint32, values []com.VARIANT) ([]int32, error) {
	return p.syncIO.Write(serverHandles, values)
}

func (p *comGroupProvider) AsyncRead(serverHandles []uint32, transactionID uint32) (uint32, []int32, error) {
	return p.asyncIO2.Read(serverHandles, transactionID)
}

func (p *comGroupProvider) AsyncWrite(serverHandles []uint32, values []com.VARIANT, transactionID uint32) (uint32, []int32, error) {
	return p.asyncIO2.Write(serverHandles, values, transactionID)
}

func (p *comGroupProvider) AsyncRefresh(source com.OPCDATASOURCE, transactionID uint32) (uint32, error) {
	return p.asyncIO2.Refresh2(source, transactionID)
}

func (p *comGroupProvider) AsyncCancel(cancelID uint32) error {
	return p.asyncIO2.Cancel2(cancelID)
}

func (p *comGroupProvider) QueryInterface(iid *windows.GUID, ppv unsafe.Pointer) error {
	return p.groupStateMgt.IUnknown.QueryInterface(iid, ppv)
}

func (p *comGroupProvider) Release() {
	if p.groupStateMgt != nil {
		p.groupStateMgt.Release()
	}
	if p.syncIO != nil {
		p.syncIO.Release()
	}
	if p.asyncIO2 != nil {
		p.asyncIO2.Release()
	}
}

type OPCGroup struct {
	parent             *OPCGroups
	provider           serverProvider
	groupProvider      groupProvider
	clientGroupHandle  uint32
	serverGroupHandle  uint32
	groupName          string
	revisedUpdateRate  uint32
	items              *OPCItems
	callbackLock       sync.Mutex
	container          *com.IConnectionPointContainer
	point              *com.IConnectionPoint
	event              *DataEventReceiver
	cookie             uint32
	ctx                context.Context
	cancel             context.CancelFunc
	dataChangeList     []chan *DataChangeCallBackData
	readCompleteList   []chan *ReadCompleteCallBackData
	writeCompleteList  []chan *WriteCompleteCallBackData
	cancelCompleteList []chan *CancelCompleteCallBackData
}

func NewOPCGroup(
	opcGroups *OPCGroups,
	iUnknown *com.IUnknown,
	clientGroupHandle uint32,
	serverGroupHandle uint32,
	groupName string,
	revisedUpdateRate uint32,
) (*OPCGroup, error) {
	if iUnknown == nil {
		return nil, errors.New("nil interface")
	}
	var iUnknownSyncIO *com.IUnknown
	err := iUnknown.QueryInterface(&com.IID_IOPCSyncIO, unsafe.Pointer(&iUnknownSyncIO))
	if err != nil {
		return nil, NewOPCWrapperError("query interface IOPCSyncIO", err)
	}
	var iUnknownAsyncIO2 *com.IUnknown
	err = iUnknown.QueryInterface(&com.IID_IOPCAsyncIO2, unsafe.Pointer(&iUnknownAsyncIO2))
	if err != nil {
		iUnknownSyncIO.Release()
		return nil, NewOPCWrapperError("query interface IOPCAsyncIO2", err)
	}
	var iUnknownItemMgt *com.IUnknown
	err = iUnknown.QueryInterface(&com.IID_IOPCItemMgt, unsafe.Pointer(&iUnknownItemMgt))
	if err != nil {
		iUnknownSyncIO.Release()
		iUnknownAsyncIO2.Release()
		return nil, NewOPCWrapperError("query interface IOPCItemMgt", err)
	}

	o := &OPCGroup{
		parent: opcGroups,
		groupProvider: &comGroupProvider{
			groupStateMgt: &com.IOPCGroupStateMgt{IUnknown: iUnknown},
			syncIO:        &com.IOPCSyncIO{IUnknown: iUnknownSyncIO},
			asyncIO2:      &com.IOPCAsyncIO2{IUnknown: iUnknownAsyncIO2},
		},
		clientGroupHandle: clientGroupHandle,
		serverGroupHandle: serverGroupHandle,
		groupName:         groupName,
		revisedUpdateRate: revisedUpdateRate,
		provider:          opcGroups.provider,
	}
	itemMgt := &comItemMgtProvider{itemMgt: &com.IOPCItemMgt{IUnknown: iUnknownItemMgt}}
	o.items = NewOPCItems(o, itemMgt, opcGroups.provider)
	return o, nil
}

// GetParent Returns reference to the parent OPCServer object
func (g *OPCGroup) GetParent() *OPCGroups {
	if g == nil {
		return nil
	}
	return g.parent
}

// GetName Returns the name of the group
func (g *OPCGroup) GetName() string {
	if g == nil {
		return ""
	}
	return g.groupName
}

// SetName set the name of the group
func (g *OPCGroup) SetName(name string) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	err := g.groupProvider.SetName(name)
	if err != nil {
		return err
	}
	g.groupName = name
	return nil
}

// GetIsActive Returns whether the group is active
func (g *OPCGroup) GetIsActive() bool {
	if g == nil || g.groupProvider == nil {
		return false
	}
	_, b, _, _, _, _, _, _, err := g.groupProvider.GetState()
	if err != nil {
		return false
	}
	return b
}

// SetIsActive set whether the group is active
func (g *OPCGroup) SetIsActive(isActive bool) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	v := com.BoolToComBOOL(isActive)
	_, err := g.groupProvider.SetState(nil, &v, nil, nil, nil, nil)
	return err
}

// GetClientHandle get a Long value associated with the group
func (g *OPCGroup) GetClientHandle() uint32 {
	if g == nil {
		return 0
	}
	return g.clientGroupHandle
}

// SetClientHandle set a Long value associated with the group
func (g *OPCGroup) SetClientHandle(clientHandle uint32) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	_, err := g.groupProvider.SetState(nil, nil, nil, nil, nil, &clientHandle)
	if err != nil {
		return err
	}
	g.clientGroupHandle = clientHandle
	return nil
}

// GetServerHandle get the server assigned handle for the group
func (g *OPCGroup) GetServerHandle() uint32 {
	if g == nil {
		return 0
	}
	return g.serverGroupHandle
}

// GetLocaleID get the locale identifier for the group
func (g *OPCGroup) GetLocaleID() (uint32, error) {
	if g == nil || g.groupProvider == nil {
		return 0, errors.New("uninitialized group")
	}
	_, _, _, _, _, localeID, _, _, err := g.groupProvider.GetState()
	return localeID, err
}

// SetLocaleID set the locale identifier for the group
func (g *OPCGroup) SetLocaleID(id uint32) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	_, err := g.groupProvider.SetState(nil, nil, nil, nil, &id, nil)
	return err
}

// GetTimeBias This property provides the information needed to convert the time stamp on the data back to the local time of the device
func (g *OPCGroup) GetTimeBias() (int32, error) {
	if g == nil || g.groupProvider == nil {
		return 0, errors.New("uninitialized group")
	}
	_, _, _, timeBias, _, _, _, _, err := g.groupProvider.GetState()
	return timeBias, err
}

// SetTimeBias This property provides the information needed to convert the time stamp on the data back to the local time of the device
func (g *OPCGroup) SetTimeBias(timeBias int32) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	_, err := g.groupProvider.SetState(nil, nil, &timeBias, nil, nil, nil)
	return err
}

// GetDeadband A deadband is expressed as percent of full scale (legal values 0 to 100).
func (g *OPCGroup) GetDeadband() (float32, error) {
	if g == nil || g.groupProvider == nil {
		return 0, errors.New("uninitialized group")
	}
	_, _, _, _, deadband, _, _, _, err := g.groupProvider.GetState()
	return deadband, err
}

// SetDeadband A deadband is expressed as percent of full scale (legal values 0 to 100).
func (g *OPCGroup) SetDeadband(deadband float32) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	_, err := g.groupProvider.SetState(nil, nil, nil, &deadband, nil, nil)
	return err
}

// GetUpdateRate
// The fastest rate at which data change events may be fired. A slow process might
// cause data changes to fire at less than this rate, but they will never exceed this rate. Rate is in
// milliseconds. This property’s default depends on the value set in the OPCGroups Collection.
// Assigning a value to this property is a “request” for a new update rate. The server may not support
// that rate, so reading the property may result in a different rate (the server will use the closest rate it
// does support).
func (g *OPCGroup) GetUpdateRate() (uint32, error) {
	if g == nil || g.groupProvider == nil {
		return 0, errors.New("uninitialized group")
	}
	updateRate, _, _, _, _, _, _, _, err := g.groupProvider.GetState()
	return updateRate, err
}

// SetUpdateRate set the update rate
func (g *OPCGroup) SetUpdateRate(updateRate uint32) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	_, err := g.groupProvider.SetState(&updateRate, nil, nil, nil, nil, nil)
	return err
}

// OPCItems A collection of OPCItem objects
func (g *OPCGroup) OPCItems() *OPCItems {
	if g == nil {
		return nil
	}
	return g.items
}

// SyncRead reads the value, quality and timestamp information for one or more items in a group.
func (g *OPCGroup) SyncRead(source com.OPCDATASOURCE, serverHandles []uint32) ([]*com.ItemState, []error, error) {
	if g == nil || g.groupProvider == nil {
		return nil, nil, errors.New("uninitialized group")
	}
	values, errList, err := g.groupProvider.SyncRead(source, serverHandles)
	if err != nil {
		return nil, nil, err
	}

	resultErrs := make([]error, len(serverHandles))
	for i, e := range errList {
		if e < 0 {
			resultErrs[i] = g.getError(e)
		}
	}

	return values, resultErrs, nil
}

// SyncWrite Writes values to one or more items in a group
func (g *OPCGroup) SyncWrite(serverHandles []uint32, values []interface{}) ([]error, error) {
	if g == nil || g.groupProvider == nil {
		return nil, errors.New("uninitialized group")
	}
	variants := make([]com.VARIANT, len(values))
	variantWrappers := make([]*com.VariantWrapper, len(values))
	defer func() {
		for _, variant := range variantWrappers {
			if variant != nil {
				err := variant.Clear()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
	for i, v := range values {
		variant, err := com.NewVariant(v)
		if err != nil {
			return nil, err
		}
		variantWrappers[i] = variant
		variants[i] = *variant.Variant
	}
	errList, err := g.groupProvider.SyncWrite(serverHandles, variants)
	if err != nil {
		return nil, err
	}
	errs := make([]error, len(errList))
	for i, e := range errList {
		if e < 0 {
			errs[i] = g.getError(e)
		}
	}
	return errs, nil
}

// Release Releases the resources used by the group
func (g *OPCGroup) Release() {
	if g == nil {
		return
	}
	if g.event != nil {
		g.point.Unadvise(g.cookie)
		g.point.Release()
		g.container.Release()
		g.event = nil
	}
	if g.cancel != nil {
		g.cancel()
	}
	if g.items != nil {
		g.items.Release()
	}
	if g.groupProvider != nil {
		g.groupProvider.Release()
	}
}

type DataChangeCallBackData struct {
	TransID           uint32
	GroupHandle       uint32
	MasterQuality     int32
	MasterErr         error
	ItemClientHandles []uint32
	Values            []interface{}
	Qualities         []uint16
	TimeStamps        []time.Time
	Errors            []error
}

// RegisterDataChange Register to receive data change events
func (g *OPCGroup) RegisterDataChange(ch chan *DataChangeCallBackData) error {
	if g == nil {
		return errors.New("uninitialized group")
	}
	err := g.advise()
	if err != nil {
		return err
	}
	g.dataChangeList = append(g.dataChangeList, ch)
	return nil
}

// RegisterReadComplete Register to receive read complete events
func (g *OPCGroup) RegisterReadComplete(ch chan *ReadCompleteCallBackData) error {
	if g == nil {
		return errors.New("uninitialized group")
	}
	err := g.advise()
	if err != nil {
		return err
	}
	g.readCompleteList = append(g.readCompleteList, ch)
	return nil
}

// RegisterWriteComplete Register to receive write complete events
func (g *OPCGroup) RegisterWriteComplete(ch chan *WriteCompleteCallBackData) error {
	if g == nil {
		return errors.New("uninitialized group")
	}
	err := g.advise()
	if err != nil {
		return err
	}
	g.writeCompleteList = append(g.writeCompleteList, ch)
	return nil
}

// RegisterCancelComplete Register to receive cancel complete events
func (g *OPCGroup) RegisterCancelComplete(ch chan *CancelCompleteCallBackData) error {
	if g == nil {
		return errors.New("uninitialized group")
	}
	err := g.advise()
	if err != nil {
		return err
	}
	g.cancelCompleteList = append(g.cancelCompleteList, ch)
	return nil
}

type ReadCompleteCallBackData struct {
	TransID           uint32
	GroupHandle       uint32
	MasterQuality     int32
	MasterErr         error
	ItemClientHandles []uint32
	Values            []interface{}
	Qualities         []uint16
	TimeStamps        []time.Time
	Errors            []error
}

type WriteCompleteCallBackData struct {
	TransID           uint32
	GroupHandle       uint32
	MasterErr         error
	ItemClientHandles []uint32
	Errors            []error
}

type CancelCompleteCallBackData struct {
	TransID     uint32
	GroupHandle uint32
}

func (g *OPCGroup) advise() (err error) {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	g.callbackLock.Lock()
	defer g.callbackLock.Unlock()
	if g.event != nil {
		return nil
	}
	var iUnknownContainer *com.IUnknown
	err = g.groupProvider.QueryInterface(&com.IID_IConnectionPointContainer, unsafe.Pointer(&iUnknownContainer))
	if err != nil {
		return NewOPCWrapperError("query interface IConnectionPointContainer", err)
	}
	defer func() {
		if err != nil {
			iUnknownContainer.Release()
		}
	}()
	container := &com.IConnectionPointContainer{IUnknown: iUnknownContainer}
	var point *com.IConnectionPoint
	point, err = container.FindConnectionPoint(&IID_IOPCDataCallback)
	if err != nil {
		return
	}
	dataChangeCB := make(chan *CDataChangeCallBackData, 100)
	readCB := make(chan *CReadCompleteCallBackData, 100)
	writeCB := make(chan *CWriteCompleteCallBackData, 100)
	cancelCB := make(chan *CCancelCompleteCallBackData, 100)
	event := NewDataEventReceiver(dataChangeCB, readCB, writeCB, cancelCB)
	var cookie uint32
	cookie, err = point.Advise((*com.IUnknown)(unsafe.Pointer(event)))
	if err != nil {
		return
	}
	g.ctx, g.cancel = context.WithCancel(context.Background())
	go g.loop(g.ctx, dataChangeCB, readCB, writeCB, cancelCB)
	g.container = container
	g.point = point
	g.event = event
	g.cookie = cookie
	return
}

func (g *OPCGroup) loop(ctx context.Context, dataChangeCB chan *CDataChangeCallBackData, readCB chan *CReadCompleteCallBackData, writeCB chan *CWriteCompleteCallBackData, cancelCB chan *CCancelCompleteCallBackData) {
	for {
		select {
		case <-ctx.Done():
			return
		case cbData := <-dataChangeCB:
			g.fireDataChange(cbData)
		case cbData := <-readCB:
			g.fireReadComplete(cbData)
		case cbData := <-writeCB:
			g.fireWriteComplete(cbData)
		case cbData := <-cancelCB:
			g.fireCancelComplete(cbData)
		}
	}
}

func (g *OPCGroup) fireDataChange(cbData *CDataChangeCallBackData) {
	if g == nil {
		return
	}
	masterError := error(nil)
	if (cbData.MasterErr) < 0 {
		masterError = g.getError(cbData.MasterErr)
	}
	itemErrors := make([]error, len(cbData.Errors))
	for i, e := range cbData.Errors {
		if e < 0 {
			itemErrors[i] = g.getError(e)
		}
	}
	data := &DataChangeCallBackData{
		TransID:           cbData.TransID,
		GroupHandle:       cbData.GroupHandle,
		MasterQuality:     cbData.MasterQuality,
		MasterErr:         masterError,
		ItemClientHandles: cbData.ItemClientHandles,
		Values:            cbData.Values,
		Qualities:         cbData.Qualities,
		TimeStamps:        cbData.TimeStamps,
		Errors:            itemErrors,
	}
	for _, backData := range g.dataChangeList {
		select {
		case backData <- data:
		default:
		}
	}
}

func (g *OPCGroup) fireReadComplete(cbData *CReadCompleteCallBackData) {
	if g == nil {
		return
	}
	masterError := error(nil)
	if (cbData.MasterErr) < 0 {
		masterError = g.getError(cbData.MasterErr)
	}
	itemErrors := make([]error, len(cbData.Errors))
	for i, e := range cbData.Errors {
		if e < 0 {
			itemErrors[i] = g.getError(e)
		}
	}
	data := &ReadCompleteCallBackData{
		TransID:           cbData.TransID,
		GroupHandle:       cbData.GroupHandle,
		MasterQuality:     cbData.MasterQuality,
		MasterErr:         masterError,
		ItemClientHandles: cbData.ItemClientHandles,
		Values:            cbData.Values,
		Qualities:         cbData.Qualities,
		TimeStamps:        cbData.TimeStamps,
		Errors:            itemErrors,
	}
	for _, backData := range g.readCompleteList {
		select {
		case backData <- data:
		default:
		}
	}
}

func (g *OPCGroup) fireWriteComplete(cbData *CWriteCompleteCallBackData) {
	if g == nil {
		return
	}
	masterError := error(nil)
	if (cbData.MasterErr) < 0 {
		masterError = g.getError(cbData.MasterErr)
	}
	itemErrors := make([]error, len(cbData.Errors))
	for i, e := range cbData.Errors {
		if e < 0 {
			itemErrors[i] = g.getError(e)
		}
	}
	data := &WriteCompleteCallBackData{
		TransID:           cbData.TransID,
		GroupHandle:       cbData.GroupHandle,
		MasterErr:         masterError,
		ItemClientHandles: cbData.ItemClientHandles,
		Errors:            itemErrors,
	}
	for _, backData := range g.writeCompleteList {
		select {
		case backData <- data:
		default:
		}
	}
}

func (g *OPCGroup) fireCancelComplete(cbData *CCancelCompleteCallBackData) {
	if g == nil {
		return
	}
	data := &CancelCompleteCallBackData{
		TransID:     cbData.TransID,
		GroupHandle: cbData.GroupHandle,
	}
	for _, backData := range g.cancelCompleteList {
		backData <- data
	}
}

// AsyncRead Read one or more items in a group. The results are returned via the AsyncReadComplete event associated with the OPCGroup object.
func (g *OPCGroup) AsyncRead(
	serverHandles []uint32,
	clientTransactionID uint32,
) (cancelID uint32, errs []error, err error) {
	if g == nil || g.groupProvider == nil {
		return 0, nil, errors.New("uninitialized group")
	}
	var es []int32
	cancelID, es, err = g.groupProvider.AsyncRead(
		serverHandles,
		clientTransactionID,
	)
	if err != nil {
		return
	}
	errs = make([]error, len(es))
	for i, e := range es {
		if e < 0 {
			errs[i] = g.getError(e)
		}
	}
	return
}

// AsyncWrite Write one or more items in a group. The results are returned via the AsyncWriteComplete event associated with the OPCGroup object.
func (g *OPCGroup) AsyncWrite(
	serverHandles []uint32,
	values []interface{},
	clientTransactionID uint32,
) (cancelID uint32, errs []error, err error) {
	if g == nil || g.groupProvider == nil {
		return 0, nil, errors.New("uninitialized group")
	}
	variants := make([]com.VARIANT, len(values))
	variantWrappers := make([]*com.VariantWrapper, len(values))

	defer func() {
		for _, v := range variants {
			v.Clear()
		}
	}()
	for i, v := range values {
		variant, err := com.NewVariant(v)
		if err != nil {
			return 0, nil, err
		}
		variantWrappers[i] = variant
		variants[i] = *variant.Variant
	}
	var es []int32
	cancelID, es, err = g.groupProvider.AsyncWrite(
		serverHandles,
		variants,
		clientTransactionID,
	)
	if err != nil {
		return
	}
	errs = make([]error, len(es))
	for i, e := range es {
		if e < 0 {
			errs[i] = g.getError(e)
		}
	}
	return
}

// AsyncRefresh Generate an event for all active items in the group (whether they have changed or not). Inactive
// items are not included in the callback. The results are returned via the DataChange event
// associated with the OPCGroup object.
func (g *OPCGroup) AsyncRefresh(
	source com.OPCDATASOURCE,
	clientTransactionID uint32,
) (cancelID uint32, err error) {
	if g == nil || g.groupProvider == nil {
		return 0, errors.New("uninitialized group")
	}
	cancelID, err = g.groupProvider.AsyncRefresh(
		source,
		clientTransactionID,
	)
	return
}

// AsyncCancel Request that the server cancel an outstanding transaction. An AsyncCancelComplete event will
// occur indicating whether or not the cancel succeeded.
func (g *OPCGroup) AsyncCancel(cancelID uint32) error {
	if g == nil || g.groupProvider == nil {
		return errors.New("uninitialized group")
	}
	return g.groupProvider.AsyncCancel(cancelID)
}

func (g *OPCGroup) getError(errorCode int32) error {
	if g == nil || g.provider == nil {
		return &OPCError{ErrorCode: errorCode, ErrorMessage: "uninitialized common interface"}
	}
	errStr, _ := g.provider.GetErrorString(uint32(errorCode))
	return &OPCError{
		ErrorCode:    errorCode,
		ErrorMessage: errStr,
	}
}
