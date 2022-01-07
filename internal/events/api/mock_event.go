// Code generated by MockGen. DO NOT EDIT.
// Source: event.go

// Package api is a generated GoMock package.
package api

import (
	context "context"
	reflect "reflect"
	time "time"

	strfmt "github.com/go-openapi/strfmt"
	gomock "github.com/golang/mock/gomock"
	common "github.com/openshift/assisted-service/internal/common"
)

// MockSender is a mock of Sender interface.
type MockSender struct {
	ctrl     *gomock.Controller
	recorder *MockSenderMockRecorder
}

// MockSenderMockRecorder is the mock recorder for MockSender.
type MockSenderMockRecorder struct {
	mock *MockSender
}

// NewMockSender creates a new mock instance.
func NewMockSender(ctrl *gomock.Controller) *MockSender {
	mock := &MockSender{ctrl: ctrl}
	mock.recorder = &MockSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSender) EXPECT() *MockSenderMockRecorder {
	return m.recorder
}

// AddEvent mocks base method.
func (m *MockSender) AddEvent(ctx context.Context, clusterID strfmt.UUID, hostID *strfmt.UUID, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AddEvent", varargs...)
}

// AddEvent indicates an expected call of AddEvent.
func (mr *MockSenderMockRecorder) AddEvent(ctx, clusterID, hostID, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEvent", reflect.TypeOf((*MockSender)(nil).AddEvent), varargs...)
}

// AddMetricsEvent mocks base method.
func (m *MockSender) AddMetricsEvent(ctx context.Context, clusterID strfmt.UUID, hostID *strfmt.UUID, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AddMetricsEvent", varargs...)
}

// AddMetricsEvent indicates an expected call of AddMetricsEvent.
func (mr *MockSenderMockRecorder) AddMetricsEvent(ctx, clusterID, hostID, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMetricsEvent", reflect.TypeOf((*MockSender)(nil).AddMetricsEvent), varargs...)
}

// SendClusterEvent mocks base method.
func (m *MockSender) SendClusterEvent(ctx context.Context, event ClusterEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendClusterEvent", ctx, event)
}

// SendClusterEvent indicates an expected call of SendClusterEvent.
func (mr *MockSenderMockRecorder) SendClusterEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendClusterEvent", reflect.TypeOf((*MockSender)(nil).SendClusterEvent), ctx, event)
}

// SendClusterEventAtTime mocks base method.
func (m *MockSender) SendClusterEventAtTime(ctx context.Context, event ClusterEvent, eventTime time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendClusterEventAtTime", ctx, event, eventTime)
}

// SendClusterEventAtTime indicates an expected call of SendClusterEventAtTime.
func (mr *MockSenderMockRecorder) SendClusterEventAtTime(ctx, event, eventTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendClusterEventAtTime", reflect.TypeOf((*MockSender)(nil).SendClusterEventAtTime), ctx, event, eventTime)
}

// SendHostEvent mocks base method.
func (m *MockSender) SendHostEvent(ctx context.Context, event HostEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendHostEvent", ctx, event)
}

// SendHostEvent indicates an expected call of SendHostEvent.
func (mr *MockSenderMockRecorder) SendHostEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHostEvent", reflect.TypeOf((*MockSender)(nil).SendHostEvent), ctx, event)
}

// SendHostEventAtTime mocks base method.
func (m *MockSender) SendHostEventAtTime(ctx context.Context, event HostEvent, eventTime time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendHostEventAtTime", ctx, event, eventTime)
}

// SendHostEventAtTime indicates an expected call of SendHostEventAtTime.
func (mr *MockSenderMockRecorder) SendHostEventAtTime(ctx, event, eventTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHostEventAtTime", reflect.TypeOf((*MockSender)(nil).SendHostEventAtTime), ctx, event, eventTime)
}

// SendInfraEnvEvent mocks base method.
func (m *MockSender) SendInfraEnvEvent(ctx context.Context, event InfraEnvEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendInfraEnvEvent", ctx, event)
}

// SendInfraEnvEvent indicates an expected call of SendInfraEnvEvent.
func (mr *MockSenderMockRecorder) SendInfraEnvEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendInfraEnvEvent", reflect.TypeOf((*MockSender)(nil).SendInfraEnvEvent), ctx, event)
}

// SendInfraEnvEventAtTime mocks base method.
func (m *MockSender) SendInfraEnvEventAtTime(ctx context.Context, event InfraEnvEvent, eventTime time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendInfraEnvEventAtTime", ctx, event, eventTime)
}

// SendInfraEnvEventAtTime indicates an expected call of SendInfraEnvEventAtTime.
func (mr *MockSenderMockRecorder) SendInfraEnvEventAtTime(ctx, event, eventTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendInfraEnvEventAtTime", reflect.TypeOf((*MockSender)(nil).SendInfraEnvEventAtTime), ctx, event, eventTime)
}

// V2AddEvent mocks base method.
func (m *MockSender) V2AddEvent(ctx context.Context, clusterID, hostID, infraEnvID *strfmt.UUID, name, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "V2AddEvent", varargs...)
}

// V2AddEvent indicates an expected call of V2AddEvent.
func (mr *MockSenderMockRecorder) V2AddEvent(ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "V2AddEvent", reflect.TypeOf((*MockSender)(nil).V2AddEvent), varargs...)
}

// V2AddMetricsEvent mocks base method.
func (m *MockSender) V2AddMetricsEvent(ctx context.Context, clusterID, hostID, infraEnvID *strfmt.UUID, name, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "V2AddMetricsEvent", varargs...)
}

// V2AddMetricsEvent indicates an expected call of V2AddMetricsEvent.
func (mr *MockSenderMockRecorder) V2AddMetricsEvent(ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "V2AddMetricsEvent", reflect.TypeOf((*MockSender)(nil).V2AddMetricsEvent), varargs...)
}

// MockHandler is a mock of Handler interface.
type MockHandler struct {
	ctrl     *gomock.Controller
	recorder *MockHandlerMockRecorder
}

// MockHandlerMockRecorder is the mock recorder for MockHandler.
type MockHandlerMockRecorder struct {
	mock *MockHandler
}

// NewMockHandler creates a new mock instance.
func NewMockHandler(ctrl *gomock.Controller) *MockHandler {
	mock := &MockHandler{ctrl: ctrl}
	mock.recorder = &MockHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHandler) EXPECT() *MockHandlerMockRecorder {
	return m.recorder
}

// AddEvent mocks base method.
func (m *MockHandler) AddEvent(ctx context.Context, clusterID strfmt.UUID, hostID *strfmt.UUID, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AddEvent", varargs...)
}

// AddEvent indicates an expected call of AddEvent.
func (mr *MockHandlerMockRecorder) AddEvent(ctx, clusterID, hostID, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEvent", reflect.TypeOf((*MockHandler)(nil).AddEvent), varargs...)
}

// AddMetricsEvent mocks base method.
func (m *MockHandler) AddMetricsEvent(ctx context.Context, clusterID strfmt.UUID, hostID *strfmt.UUID, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AddMetricsEvent", varargs...)
}

// AddMetricsEvent indicates an expected call of AddMetricsEvent.
func (mr *MockHandlerMockRecorder) AddMetricsEvent(ctx, clusterID, hostID, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMetricsEvent", reflect.TypeOf((*MockHandler)(nil).AddMetricsEvent), varargs...)
}

// GetEvents mocks base method.
func (m *MockHandler) GetEvents(clusterID strfmt.UUID, hostID *strfmt.UUID, categories ...string) ([]*common.Event, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{clusterID, hostID}
	for _, a := range categories {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEvents", varargs...)
	ret0, _ := ret[0].([]*common.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEvents indicates an expected call of GetEvents.
func (mr *MockHandlerMockRecorder) GetEvents(clusterID, hostID interface{}, categories ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{clusterID, hostID}, categories...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEvents", reflect.TypeOf((*MockHandler)(nil).GetEvents), varargs...)
}

// SendClusterEvent mocks base method.
func (m *MockHandler) SendClusterEvent(ctx context.Context, event ClusterEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendClusterEvent", ctx, event)
}

// SendClusterEvent indicates an expected call of SendClusterEvent.
func (mr *MockHandlerMockRecorder) SendClusterEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendClusterEvent", reflect.TypeOf((*MockHandler)(nil).SendClusterEvent), ctx, event)
}

// SendClusterEventAtTime mocks base method.
func (m *MockHandler) SendClusterEventAtTime(ctx context.Context, event ClusterEvent, eventTime time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendClusterEventAtTime", ctx, event, eventTime)
}

// SendClusterEventAtTime indicates an expected call of SendClusterEventAtTime.
func (mr *MockHandlerMockRecorder) SendClusterEventAtTime(ctx, event, eventTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendClusterEventAtTime", reflect.TypeOf((*MockHandler)(nil).SendClusterEventAtTime), ctx, event, eventTime)
}

// SendHostEvent mocks base method.
func (m *MockHandler) SendHostEvent(ctx context.Context, event HostEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendHostEvent", ctx, event)
}

// SendHostEvent indicates an expected call of SendHostEvent.
func (mr *MockHandlerMockRecorder) SendHostEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHostEvent", reflect.TypeOf((*MockHandler)(nil).SendHostEvent), ctx, event)
}

// SendHostEventAtTime mocks base method.
func (m *MockHandler) SendHostEventAtTime(ctx context.Context, event HostEvent, eventTime time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendHostEventAtTime", ctx, event, eventTime)
}

// SendHostEventAtTime indicates an expected call of SendHostEventAtTime.
func (mr *MockHandlerMockRecorder) SendHostEventAtTime(ctx, event, eventTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHostEventAtTime", reflect.TypeOf((*MockHandler)(nil).SendHostEventAtTime), ctx, event, eventTime)
}

// SendInfraEnvEvent mocks base method.
func (m *MockHandler) SendInfraEnvEvent(ctx context.Context, event InfraEnvEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendInfraEnvEvent", ctx, event)
}

// SendInfraEnvEvent indicates an expected call of SendInfraEnvEvent.
func (mr *MockHandlerMockRecorder) SendInfraEnvEvent(ctx, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendInfraEnvEvent", reflect.TypeOf((*MockHandler)(nil).SendInfraEnvEvent), ctx, event)
}

// SendInfraEnvEventAtTime mocks base method.
func (m *MockHandler) SendInfraEnvEventAtTime(ctx context.Context, event InfraEnvEvent, eventTime time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendInfraEnvEventAtTime", ctx, event, eventTime)
}

// SendInfraEnvEventAtTime indicates an expected call of SendInfraEnvEventAtTime.
func (mr *MockHandlerMockRecorder) SendInfraEnvEventAtTime(ctx, event, eventTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendInfraEnvEventAtTime", reflect.TypeOf((*MockHandler)(nil).SendInfraEnvEventAtTime), ctx, event, eventTime)
}

// V2AddEvent mocks base method.
func (m *MockHandler) V2AddEvent(ctx context.Context, clusterID, hostID, infraEnvID *strfmt.UUID, name, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "V2AddEvent", varargs...)
}

// V2AddEvent indicates an expected call of V2AddEvent.
func (mr *MockHandlerMockRecorder) V2AddEvent(ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "V2AddEvent", reflect.TypeOf((*MockHandler)(nil).V2AddEvent), varargs...)
}

// V2AddMetricsEvent mocks base method.
func (m *MockHandler) V2AddMetricsEvent(ctx context.Context, clusterID, hostID, infraEnvID *strfmt.UUID, name, severity, msg string, eventTime time.Time, props ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}
	for _, a := range props {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "V2AddMetricsEvent", varargs...)
}

// V2AddMetricsEvent indicates an expected call of V2AddMetricsEvent.
func (mr *MockHandlerMockRecorder) V2AddMetricsEvent(ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime interface{}, props ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, infraEnvID, name, severity, msg, eventTime}, props...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "V2AddMetricsEvent", reflect.TypeOf((*MockHandler)(nil).V2AddMetricsEvent), varargs...)
}

// V2GetEvents mocks base method.
func (m *MockHandler) V2GetEvents(ctx context.Context, clusterID, hostID, infraEnvID *strfmt.UUID, categories ...string) ([]*common.Event, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, clusterID, hostID, infraEnvID}
	for _, a := range categories {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "V2GetEvents", varargs...)
	ret0, _ := ret[0].([]*common.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// V2GetEvents indicates an expected call of V2GetEvents.
func (mr *MockHandlerMockRecorder) V2GetEvents(ctx, clusterID, hostID, infraEnvID interface{}, categories ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, clusterID, hostID, infraEnvID}, categories...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "V2GetEvents", reflect.TypeOf((*MockHandler)(nil).V2GetEvents), varargs...)
}

// MockBaseEvent is a mock of BaseEvent interface.
type MockBaseEvent struct {
	ctrl     *gomock.Controller
	recorder *MockBaseEventMockRecorder
}

// MockBaseEventMockRecorder is the mock recorder for MockBaseEvent.
type MockBaseEventMockRecorder struct {
	mock *MockBaseEvent
}

// NewMockBaseEvent creates a new mock instance.
func NewMockBaseEvent(ctrl *gomock.Controller) *MockBaseEvent {
	mock := &MockBaseEvent{ctrl: ctrl}
	mock.recorder = &MockBaseEventMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBaseEvent) EXPECT() *MockBaseEventMockRecorder {
	return m.recorder
}

// FormatMessage mocks base method.
func (m *MockBaseEvent) FormatMessage() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FormatMessage")
	ret0, _ := ret[0].(string)
	return ret0
}

// FormatMessage indicates an expected call of FormatMessage.
func (mr *MockBaseEventMockRecorder) FormatMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FormatMessage", reflect.TypeOf((*MockBaseEvent)(nil).FormatMessage))
}

// GetName mocks base method.
func (m *MockBaseEvent) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockBaseEventMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockBaseEvent)(nil).GetName))
}

// GetSeverity mocks base method.
func (m *MockBaseEvent) GetSeverity() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSeverity")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSeverity indicates an expected call of GetSeverity.
func (mr *MockBaseEventMockRecorder) GetSeverity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSeverity", reflect.TypeOf((*MockBaseEvent)(nil).GetSeverity))
}

// MockClusterEvent is a mock of ClusterEvent interface.
type MockClusterEvent struct {
	ctrl     *gomock.Controller
	recorder *MockClusterEventMockRecorder
}

// MockClusterEventMockRecorder is the mock recorder for MockClusterEvent.
type MockClusterEventMockRecorder struct {
	mock *MockClusterEvent
}

// NewMockClusterEvent creates a new mock instance.
func NewMockClusterEvent(ctrl *gomock.Controller) *MockClusterEvent {
	mock := &MockClusterEvent{ctrl: ctrl}
	mock.recorder = &MockClusterEventMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClusterEvent) EXPECT() *MockClusterEventMockRecorder {
	return m.recorder
}

// FormatMessage mocks base method.
func (m *MockClusterEvent) FormatMessage() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FormatMessage")
	ret0, _ := ret[0].(string)
	return ret0
}

// FormatMessage indicates an expected call of FormatMessage.
func (mr *MockClusterEventMockRecorder) FormatMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FormatMessage", reflect.TypeOf((*MockClusterEvent)(nil).FormatMessage))
}

// GetClusterId mocks base method.
func (m *MockClusterEvent) GetClusterId() strfmt.UUID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterId")
	ret0, _ := ret[0].(strfmt.UUID)
	return ret0
}

// GetClusterId indicates an expected call of GetClusterId.
func (mr *MockClusterEventMockRecorder) GetClusterId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterId", reflect.TypeOf((*MockClusterEvent)(nil).GetClusterId))
}

// GetName mocks base method.
func (m *MockClusterEvent) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockClusterEventMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockClusterEvent)(nil).GetName))
}

// GetSeverity mocks base method.
func (m *MockClusterEvent) GetSeverity() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSeverity")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSeverity indicates an expected call of GetSeverity.
func (mr *MockClusterEventMockRecorder) GetSeverity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSeverity", reflect.TypeOf((*MockClusterEvent)(nil).GetSeverity))
}

// MockHostEvent is a mock of HostEvent interface.
type MockHostEvent struct {
	ctrl     *gomock.Controller
	recorder *MockHostEventMockRecorder
}

// MockHostEventMockRecorder is the mock recorder for MockHostEvent.
type MockHostEventMockRecorder struct {
	mock *MockHostEvent
}

// NewMockHostEvent creates a new mock instance.
func NewMockHostEvent(ctrl *gomock.Controller) *MockHostEvent {
	mock := &MockHostEvent{ctrl: ctrl}
	mock.recorder = &MockHostEventMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHostEvent) EXPECT() *MockHostEventMockRecorder {
	return m.recorder
}

// FormatMessage mocks base method.
func (m *MockHostEvent) FormatMessage() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FormatMessage")
	ret0, _ := ret[0].(string)
	return ret0
}

// FormatMessage indicates an expected call of FormatMessage.
func (mr *MockHostEventMockRecorder) FormatMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FormatMessage", reflect.TypeOf((*MockHostEvent)(nil).FormatMessage))
}

// GetClusterId mocks base method.
func (m *MockHostEvent) GetClusterId() *strfmt.UUID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterId")
	ret0, _ := ret[0].(*strfmt.UUID)
	return ret0
}

// GetClusterId indicates an expected call of GetClusterId.
func (mr *MockHostEventMockRecorder) GetClusterId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterId", reflect.TypeOf((*MockHostEvent)(nil).GetClusterId))
}

// GetHostId mocks base method.
func (m *MockHostEvent) GetHostId() strfmt.UUID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHostId")
	ret0, _ := ret[0].(strfmt.UUID)
	return ret0
}

// GetHostId indicates an expected call of GetHostId.
func (mr *MockHostEventMockRecorder) GetHostId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHostId", reflect.TypeOf((*MockHostEvent)(nil).GetHostId))
}

// GetInfraEnvId mocks base method.
func (m *MockHostEvent) GetInfraEnvId() strfmt.UUID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfraEnvId")
	ret0, _ := ret[0].(strfmt.UUID)
	return ret0
}

// GetInfraEnvId indicates an expected call of GetInfraEnvId.
func (mr *MockHostEventMockRecorder) GetInfraEnvId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfraEnvId", reflect.TypeOf((*MockHostEvent)(nil).GetInfraEnvId))
}

// GetName mocks base method.
func (m *MockHostEvent) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockHostEventMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockHostEvent)(nil).GetName))
}

// GetSeverity mocks base method.
func (m *MockHostEvent) GetSeverity() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSeverity")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSeverity indicates an expected call of GetSeverity.
func (mr *MockHostEventMockRecorder) GetSeverity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSeverity", reflect.TypeOf((*MockHostEvent)(nil).GetSeverity))
}

// MockInfraEnvEvent is a mock of InfraEnvEvent interface.
type MockInfraEnvEvent struct {
	ctrl     *gomock.Controller
	recorder *MockInfraEnvEventMockRecorder
}

// MockInfraEnvEventMockRecorder is the mock recorder for MockInfraEnvEvent.
type MockInfraEnvEventMockRecorder struct {
	mock *MockInfraEnvEvent
}

// NewMockInfraEnvEvent creates a new mock instance.
func NewMockInfraEnvEvent(ctrl *gomock.Controller) *MockInfraEnvEvent {
	mock := &MockInfraEnvEvent{ctrl: ctrl}
	mock.recorder = &MockInfraEnvEventMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInfraEnvEvent) EXPECT() *MockInfraEnvEventMockRecorder {
	return m.recorder
}

// FormatMessage mocks base method.
func (m *MockInfraEnvEvent) FormatMessage() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FormatMessage")
	ret0, _ := ret[0].(string)
	return ret0
}

// FormatMessage indicates an expected call of FormatMessage.
func (mr *MockInfraEnvEventMockRecorder) FormatMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FormatMessage", reflect.TypeOf((*MockInfraEnvEvent)(nil).FormatMessage))
}

// GetClusterId mocks base method.
func (m *MockInfraEnvEvent) GetClusterId() *strfmt.UUID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterId")
	ret0, _ := ret[0].(*strfmt.UUID)
	return ret0
}

// GetClusterId indicates an expected call of GetClusterId.
func (mr *MockInfraEnvEventMockRecorder) GetClusterId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterId", reflect.TypeOf((*MockInfraEnvEvent)(nil).GetClusterId))
}

// GetInfraEnvId mocks base method.
func (m *MockInfraEnvEvent) GetInfraEnvId() strfmt.UUID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfraEnvId")
	ret0, _ := ret[0].(strfmt.UUID)
	return ret0
}

// GetInfraEnvId indicates an expected call of GetInfraEnvId.
func (mr *MockInfraEnvEventMockRecorder) GetInfraEnvId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfraEnvId", reflect.TypeOf((*MockInfraEnvEvent)(nil).GetInfraEnvId))
}

// GetName mocks base method.
func (m *MockInfraEnvEvent) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockInfraEnvEventMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockInfraEnvEvent)(nil).GetName))
}

// GetSeverity mocks base method.
func (m *MockInfraEnvEvent) GetSeverity() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSeverity")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSeverity indicates an expected call of GetSeverity.
func (mr *MockInfraEnvEventMockRecorder) GetSeverity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSeverity", reflect.TypeOf((*MockInfraEnvEvent)(nil).GetSeverity))
}

// MockInfoEvent is a mock of InfoEvent interface.
type MockInfoEvent struct {
	ctrl     *gomock.Controller
	recorder *MockInfoEventMockRecorder
}

// MockInfoEventMockRecorder is the mock recorder for MockInfoEvent.
type MockInfoEventMockRecorder struct {
	mock *MockInfoEvent
}

// NewMockInfoEvent creates a new mock instance.
func NewMockInfoEvent(ctrl *gomock.Controller) *MockInfoEvent {
	mock := &MockInfoEvent{ctrl: ctrl}
	mock.recorder = &MockInfoEventMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInfoEvent) EXPECT() *MockInfoEventMockRecorder {
	return m.recorder
}

// FormatMessage mocks base method.
func (m *MockInfoEvent) FormatMessage() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FormatMessage")
	ret0, _ := ret[0].(string)
	return ret0
}

// FormatMessage indicates an expected call of FormatMessage.
func (mr *MockInfoEventMockRecorder) FormatMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FormatMessage", reflect.TypeOf((*MockInfoEvent)(nil).FormatMessage))
}

// GetInfo mocks base method.
func (m *MockInfoEvent) GetInfo() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInfo")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetInfo indicates an expected call of GetInfo.
func (mr *MockInfoEventMockRecorder) GetInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInfo", reflect.TypeOf((*MockInfoEvent)(nil).GetInfo))
}

// GetName mocks base method.
func (m *MockInfoEvent) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockInfoEventMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockInfoEvent)(nil).GetName))
}

// GetSeverity mocks base method.
func (m *MockInfoEvent) GetSeverity() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSeverity")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSeverity indicates an expected call of GetSeverity.
func (mr *MockInfoEventMockRecorder) GetSeverity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSeverity", reflect.TypeOf((*MockInfoEvent)(nil).GetSeverity))
}
