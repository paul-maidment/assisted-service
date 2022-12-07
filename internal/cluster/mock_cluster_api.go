// Code generated by MockGen. DO NOT EDIT.
// Source: cluster.go

// Package cluster is a generated GoMock package.
package cluster

import (
	context "context"
	reflect "reflect"

	strfmt "github.com/go-openapi/strfmt"
	gomock "github.com/golang/mock/gomock"
	common "github.com/openshift/assisted-service/internal/common"
	s3wrapper "github.com/openshift/assisted-service/pkg/s3wrapper"
	gorm "gorm.io/gorm"
	types "k8s.io/apimachinery/pkg/types"
)

// MockRegistrationAPI is a mock of RegistrationAPI interface.
type MockRegistrationAPI struct {
	ctrl     *gomock.Controller
	recorder *MockRegistrationAPIMockRecorder
}

// MockRegistrationAPIMockRecorder is the mock recorder for MockRegistrationAPI.
type MockRegistrationAPIMockRecorder struct {
	mock *MockRegistrationAPI
}

// NewMockRegistrationAPI creates a new mock instance.
func NewMockRegistrationAPI(ctrl *gomock.Controller) *MockRegistrationAPI {
	mock := &MockRegistrationAPI{ctrl: ctrl}
	mock.recorder = &MockRegistrationAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegistrationAPI) EXPECT() *MockRegistrationAPIMockRecorder {
	return m.recorder
}

// DeregisterCluster mocks base method.
func (m *MockRegistrationAPI) DeregisterCluster(ctx context.Context, c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeregisterCluster", ctx, c)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeregisterCluster indicates an expected call of DeregisterCluster.
func (mr *MockRegistrationAPIMockRecorder) DeregisterCluster(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeregisterCluster", reflect.TypeOf((*MockRegistrationAPI)(nil).DeregisterCluster), ctx, c)
}

// RegisterAddHostsCluster mocks base method.
func (m *MockRegistrationAPI) RegisterAddHostsCluster(ctx context.Context, c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterAddHostsCluster", ctx, c)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterAddHostsCluster indicates an expected call of RegisterAddHostsCluster.
func (mr *MockRegistrationAPIMockRecorder) RegisterAddHostsCluster(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterAddHostsCluster", reflect.TypeOf((*MockRegistrationAPI)(nil).RegisterAddHostsCluster), ctx, c)
}

// RegisterAddHostsOCPCluster mocks base method.
func (m *MockRegistrationAPI) RegisterAddHostsOCPCluster(c *common.Cluster, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterAddHostsOCPCluster", c, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterAddHostsOCPCluster indicates an expected call of RegisterAddHostsOCPCluster.
func (mr *MockRegistrationAPIMockRecorder) RegisterAddHostsOCPCluster(c, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterAddHostsOCPCluster", reflect.TypeOf((*MockRegistrationAPI)(nil).RegisterAddHostsOCPCluster), c, db)
}

// RegisterCluster mocks base method.
func (m *MockRegistrationAPI) RegisterCluster(ctx context.Context, c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterCluster", ctx, c)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterCluster indicates an expected call of RegisterCluster.
func (mr *MockRegistrationAPIMockRecorder) RegisterCluster(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterCluster", reflect.TypeOf((*MockRegistrationAPI)(nil).RegisterCluster), ctx, c)
}

// MockInstallationAPI is a mock of InstallationAPI interface.
type MockInstallationAPI struct {
	ctrl     *gomock.Controller
	recorder *MockInstallationAPIMockRecorder
}

// MockInstallationAPIMockRecorder is the mock recorder for MockInstallationAPI.
type MockInstallationAPIMockRecorder struct {
	mock *MockInstallationAPI
}

// NewMockInstallationAPI creates a new mock instance.
func NewMockInstallationAPI(ctrl *gomock.Controller) *MockInstallationAPI {
	mock := &MockInstallationAPI{ctrl: ctrl}
	mock.recorder = &MockInstallationAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInstallationAPI) EXPECT() *MockInstallationAPIMockRecorder {
	return m.recorder
}

// GetMasterNodesIds mocks base method.
func (m *MockInstallationAPI) GetMasterNodesIds(ctx context.Context, c *common.Cluster, db *gorm.DB) ([]*strfmt.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMasterNodesIds", ctx, c, db)
	ret0, _ := ret[0].([]*strfmt.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMasterNodesIds indicates an expected call of GetMasterNodesIds.
func (mr *MockInstallationAPIMockRecorder) GetMasterNodesIds(ctx, c, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMasterNodesIds", reflect.TypeOf((*MockInstallationAPI)(nil).GetMasterNodesIds), ctx, c, db)
}

// MockProgressAPI is a mock of ProgressAPI interface.
type MockProgressAPI struct {
	ctrl     *gomock.Controller
	recorder *MockProgressAPIMockRecorder
}

// MockProgressAPIMockRecorder is the mock recorder for MockProgressAPI.
type MockProgressAPIMockRecorder struct {
	mock *MockProgressAPI
}

// NewMockProgressAPI creates a new mock instance.
func NewMockProgressAPI(ctrl *gomock.Controller) *MockProgressAPI {
	mock := &MockProgressAPI{ctrl: ctrl}
	mock.recorder = &MockProgressAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProgressAPI) EXPECT() *MockProgressAPIMockRecorder {
	return m.recorder
}

// UpdateFinalizingProgress mocks base method.
func (m *MockProgressAPI) UpdateFinalizingProgress(ctx context.Context, db *gorm.DB, clusterID strfmt.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFinalizingProgress", ctx, db, clusterID)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFinalizingProgress indicates an expected call of UpdateFinalizingProgress.
func (mr *MockProgressAPIMockRecorder) UpdateFinalizingProgress(ctx, db, clusterID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFinalizingProgress", reflect.TypeOf((*MockProgressAPI)(nil).UpdateFinalizingProgress), ctx, db, clusterID)
}

// UpdateInstallProgress mocks base method.
func (m *MockProgressAPI) UpdateInstallProgress(ctx context.Context, clusterID strfmt.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateInstallProgress", ctx, clusterID)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateInstallProgress indicates an expected call of UpdateInstallProgress.
func (mr *MockProgressAPIMockRecorder) UpdateInstallProgress(ctx, clusterID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateInstallProgress", reflect.TypeOf((*MockProgressAPI)(nil).UpdateInstallProgress), ctx, clusterID)
}

// MockAPI is a mock of API interface.
type MockAPI struct {
	ctrl     *gomock.Controller
	recorder *MockAPIMockRecorder
}

// MockAPIMockRecorder is the mock recorder for MockAPI.
type MockAPIMockRecorder struct {
	mock *MockAPI
}

// NewMockAPI creates a new mock instance.
func NewMockAPI(ctrl *gomock.Controller) *MockAPI {
	mock := &MockAPI{ctrl: ctrl}
	mock.recorder = &MockAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAPI) EXPECT() *MockAPIMockRecorder {
	return m.recorder
}

// AcceptRegistration mocks base method.
func (m *MockAPI) AcceptRegistration(c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AcceptRegistration", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// AcceptRegistration indicates an expected call of AcceptRegistration.
func (mr *MockAPIMockRecorder) AcceptRegistration(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptRegistration", reflect.TypeOf((*MockAPI)(nil).AcceptRegistration), c)
}

// CancelInstallation mocks base method.
func (m *MockAPI) CancelInstallation(ctx context.Context, c *common.Cluster, reason string, db *gorm.DB) *common.ApiErrorResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CancelInstallation", ctx, c, reason, db)
	ret0, _ := ret[0].(*common.ApiErrorResponse)
	return ret0
}

// CancelInstallation indicates an expected call of CancelInstallation.
func (mr *MockAPIMockRecorder) CancelInstallation(ctx, c, reason, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelInstallation", reflect.TypeOf((*MockAPI)(nil).CancelInstallation), ctx, c, reason, db)
}

// ClusterMonitoring mocks base method.
func (m *MockAPI) ClusterMonitoring() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ClusterMonitoring")
}

// ClusterMonitoring indicates an expected call of ClusterMonitoring.
func (mr *MockAPIMockRecorder) ClusterMonitoring() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClusterMonitoring", reflect.TypeOf((*MockAPI)(nil).ClusterMonitoring))
}

// CompleteInstallation mocks base method.
func (m *MockAPI) CompleteInstallation(ctx context.Context, db *gorm.DB, cluster *common.Cluster, successfullyFinished bool, reason string) (*common.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompleteInstallation", ctx, db, cluster, successfullyFinished, reason)
	ret0, _ := ret[0].(*common.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CompleteInstallation indicates an expected call of CompleteInstallation.
func (mr *MockAPIMockRecorder) CompleteInstallation(ctx, db, cluster, successfullyFinished, reason interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompleteInstallation", reflect.TypeOf((*MockAPI)(nil).CompleteInstallation), ctx, db, cluster, successfullyFinished, reason)
}

// DeleteClusterFiles mocks base method.
func (m *MockAPI) DeleteClusterFiles(ctx context.Context, c *common.Cluster, objectHandler s3wrapper.API) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteClusterFiles", ctx, c, objectHandler)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteClusterFiles indicates an expected call of DeleteClusterFiles.
func (mr *MockAPIMockRecorder) DeleteClusterFiles(ctx, c, objectHandler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteClusterFiles", reflect.TypeOf((*MockAPI)(nil).DeleteClusterFiles), ctx, c, objectHandler)
}

// DeleteClusterLogs mocks base method.
func (m *MockAPI) DeleteClusterLogs(ctx context.Context, c *common.Cluster, objectHandler s3wrapper.API) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteClusterLogs", ctx, c, objectHandler)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteClusterLogs indicates an expected call of DeleteClusterLogs.
func (mr *MockAPIMockRecorder) DeleteClusterLogs(ctx, c, objectHandler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteClusterLogs", reflect.TypeOf((*MockAPI)(nil).DeleteClusterLogs), ctx, c, objectHandler)
}

// DeregisterCluster mocks base method.
func (m *MockAPI) DeregisterCluster(ctx context.Context, c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeregisterCluster", ctx, c)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeregisterCluster indicates an expected call of DeregisterCluster.
func (mr *MockAPIMockRecorder) DeregisterCluster(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeregisterCluster", reflect.TypeOf((*MockAPI)(nil).DeregisterCluster), ctx, c)
}

// DeregisterInactiveCluster mocks base method.
func (m *MockAPI) DeregisterInactiveCluster(ctx context.Context, maxDeregisterPerInterval int, inactiveSince strfmt.DateTime) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeregisterInactiveCluster", ctx, maxDeregisterPerInterval, inactiveSince)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeregisterInactiveCluster indicates an expected call of DeregisterInactiveCluster.
func (mr *MockAPIMockRecorder) DeregisterInactiveCluster(ctx, maxDeregisterPerInterval, inactiveSince interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeregisterInactiveCluster", reflect.TypeOf((*MockAPI)(nil).DeregisterInactiveCluster), ctx, maxDeregisterPerInterval, inactiveSince)
}

// DetectAndStoreCollidingIPsForCluster mocks base method.
func (m *MockAPI) DetectAndStoreCollidingIPsForCluster(clusterID strfmt.UUID, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetectAndStoreCollidingIPsForCluster", clusterID, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// DetectAndStoreCollidingIPsForCluster indicates an expected call of DetectAndStoreCollidingIPsForCluster.
func (mr *MockAPIMockRecorder) DetectAndStoreCollidingIPsForCluster(clusterID, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetectAndStoreCollidingIPsForCluster", reflect.TypeOf((*MockAPI)(nil).DetectAndStoreCollidingIPsForCluster), clusterID, db)
}

// GenerateAdditionalManifests mocks base method.
func (m *MockAPI) GenerateAdditionalManifests(ctx context.Context, cluster *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateAdditionalManifests", ctx, cluster)
	ret0, _ := ret[0].(error)
	return ret0
}

// GenerateAdditionalManifests indicates an expected call of GenerateAdditionalManifests.
func (mr *MockAPIMockRecorder) GenerateAdditionalManifests(ctx, cluster interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateAdditionalManifests", reflect.TypeOf((*MockAPI)(nil).GenerateAdditionalManifests), ctx, cluster)
}

// GetClusterByKubeKey mocks base method.
func (m *MockAPI) GetClusterByKubeKey(key types.NamespacedName) (*common.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterByKubeKey", key)
	ret0, _ := ret[0].(*common.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClusterByKubeKey indicates an expected call of GetClusterByKubeKey.
func (mr *MockAPIMockRecorder) GetClusterByKubeKey(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterByKubeKey", reflect.TypeOf((*MockAPI)(nil).GetClusterByKubeKey), key)
}

// GetMasterNodesIds mocks base method.
func (m *MockAPI) GetMasterNodesIds(ctx context.Context, c *common.Cluster, db *gorm.DB) ([]*strfmt.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMasterNodesIds", ctx, c, db)
	ret0, _ := ret[0].([]*strfmt.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMasterNodesIds indicates an expected call of GetMasterNodesIds.
func (mr *MockAPIMockRecorder) GetMasterNodesIds(ctx, c, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMasterNodesIds", reflect.TypeOf((*MockAPI)(nil).GetMasterNodesIds), ctx, c, db)
}

// HandlePreInstallError mocks base method.
func (m *MockAPI) HandlePreInstallError(ctx context.Context, c *common.Cluster, err error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlePreInstallError", ctx, c, err)
}

// HandlePreInstallError indicates an expected call of HandlePreInstallError.
func (mr *MockAPIMockRecorder) HandlePreInstallError(ctx, c, err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlePreInstallError", reflect.TypeOf((*MockAPI)(nil).HandlePreInstallError), ctx, c, err)
}

// HandlePreInstallSuccess mocks base method.
func (m *MockAPI) HandlePreInstallSuccess(ctx context.Context, c *common.Cluster) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlePreInstallSuccess", ctx, c)
}

// HandlePreInstallSuccess indicates an expected call of HandlePreInstallSuccess.
func (mr *MockAPIMockRecorder) HandlePreInstallSuccess(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlePreInstallSuccess", reflect.TypeOf((*MockAPI)(nil).HandlePreInstallSuccess), ctx, c)
}

// IsOperatorAvailable mocks base method.
func (m *MockAPI) IsOperatorAvailable(c *common.Cluster, operatorName string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsOperatorAvailable", c, operatorName)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsOperatorAvailable indicates an expected call of IsOperatorAvailable.
func (mr *MockAPIMockRecorder) IsOperatorAvailable(c, operatorName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsOperatorAvailable", reflect.TypeOf((*MockAPI)(nil).IsOperatorAvailable), c, operatorName)
}

// IsReadyForInstallation mocks base method.
func (m *MockAPI) IsReadyForInstallation(c *common.Cluster) (bool, string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsReadyForInstallation", c)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// IsReadyForInstallation indicates an expected call of IsReadyForInstallation.
func (mr *MockAPIMockRecorder) IsReadyForInstallation(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsReadyForInstallation", reflect.TypeOf((*MockAPI)(nil).IsReadyForInstallation), c)
}

// PermanentClustersDeletion mocks base method.
func (m *MockAPI) PermanentClustersDeletion(ctx context.Context, olderThan strfmt.DateTime, objectHandler s3wrapper.API) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PermanentClustersDeletion", ctx, olderThan, objectHandler)
	ret0, _ := ret[0].(error)
	return ret0
}

// PermanentClustersDeletion indicates an expected call of PermanentClustersDeletion.
func (mr *MockAPIMockRecorder) PermanentClustersDeletion(ctx, olderThan, objectHandler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PermanentClustersDeletion", reflect.TypeOf((*MockAPI)(nil).PermanentClustersDeletion), ctx, olderThan, objectHandler)
}

// PrepareClusterLogFile mocks base method.
func (m *MockAPI) PrepareClusterLogFile(ctx context.Context, c *common.Cluster, objectHandler s3wrapper.API) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareClusterLogFile", ctx, c, objectHandler)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PrepareClusterLogFile indicates an expected call of PrepareClusterLogFile.
func (mr *MockAPIMockRecorder) PrepareClusterLogFile(ctx, c, objectHandler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareClusterLogFile", reflect.TypeOf((*MockAPI)(nil).PrepareClusterLogFile), ctx, c, objectHandler)
}

// PrepareForInstallation mocks base method.
func (m *MockAPI) PrepareForInstallation(ctx context.Context, c *common.Cluster, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareForInstallation", ctx, c, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// PrepareForInstallation indicates an expected call of PrepareForInstallation.
func (mr *MockAPIMockRecorder) PrepareForInstallation(ctx, c, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareForInstallation", reflect.TypeOf((*MockAPI)(nil).PrepareForInstallation), ctx, c, db)
}

// RefreshSchedulableMastersForcedTrue mocks base method.
func (m *MockAPI) RefreshSchedulableMastersForcedTrue(ctx context.Context, clusterID strfmt.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshSchedulableMastersForcedTrue", ctx, clusterID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RefreshSchedulableMastersForcedTrue indicates an expected call of RefreshSchedulableMastersForcedTrue.
func (mr *MockAPIMockRecorder) RefreshSchedulableMastersForcedTrue(ctx, clusterID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshSchedulableMastersForcedTrue", reflect.TypeOf((*MockAPI)(nil).RefreshSchedulableMastersForcedTrue), ctx, clusterID)
}

// RefreshStatus mocks base method.
func (m *MockAPI) RefreshStatus(ctx context.Context, c *common.Cluster, db *gorm.DB) (*common.Cluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshStatus", ctx, c, db)
	ret0, _ := ret[0].(*common.Cluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RefreshStatus indicates an expected call of RefreshStatus.
func (mr *MockAPIMockRecorder) RefreshStatus(ctx, c, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshStatus", reflect.TypeOf((*MockAPI)(nil).RefreshStatus), ctx, c, db)
}

// RegisterAddHostsCluster mocks base method.
func (m *MockAPI) RegisterAddHostsCluster(ctx context.Context, c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterAddHostsCluster", ctx, c)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterAddHostsCluster indicates an expected call of RegisterAddHostsCluster.
func (mr *MockAPIMockRecorder) RegisterAddHostsCluster(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterAddHostsCluster", reflect.TypeOf((*MockAPI)(nil).RegisterAddHostsCluster), ctx, c)
}

// RegisterAddHostsOCPCluster mocks base method.
func (m *MockAPI) RegisterAddHostsOCPCluster(c *common.Cluster, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterAddHostsOCPCluster", c, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterAddHostsOCPCluster indicates an expected call of RegisterAddHostsOCPCluster.
func (mr *MockAPIMockRecorder) RegisterAddHostsOCPCluster(c, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterAddHostsOCPCluster", reflect.TypeOf((*MockAPI)(nil).RegisterAddHostsOCPCluster), c, db)
}

// RegisterCluster mocks base method.
func (m *MockAPI) RegisterCluster(ctx context.Context, c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterCluster", ctx, c)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterCluster indicates an expected call of RegisterCluster.
func (mr *MockAPIMockRecorder) RegisterCluster(ctx, c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterCluster", reflect.TypeOf((*MockAPI)(nil).RegisterCluster), ctx, c)
}

// ResetCluster mocks base method.
func (m *MockAPI) ResetCluster(ctx context.Context, c *common.Cluster, reason string, db *gorm.DB) *common.ApiErrorResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResetCluster", ctx, c, reason, db)
	ret0, _ := ret[0].(*common.ApiErrorResponse)
	return ret0
}

// ResetCluster indicates an expected call of ResetCluster.
func (mr *MockAPIMockRecorder) ResetCluster(ctx, c, reason, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetCluster", reflect.TypeOf((*MockAPI)(nil).ResetCluster), ctx, c, reason, db)
}

// SetConnectivityMajorityGroupsForCluster mocks base method.
func (m *MockAPI) SetConnectivityMajorityGroupsForCluster(clusterID strfmt.UUID, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetConnectivityMajorityGroupsForCluster", clusterID, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetConnectivityMajorityGroupsForCluster indicates an expected call of SetConnectivityMajorityGroupsForCluster.
func (mr *MockAPIMockRecorder) SetConnectivityMajorityGroupsForCluster(clusterID, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetConnectivityMajorityGroupsForCluster", reflect.TypeOf((*MockAPI)(nil).SetConnectivityMajorityGroupsForCluster), clusterID, db)
}

// SetUploadControllerLogsAt mocks base method.
func (m *MockAPI) SetUploadControllerLogsAt(ctx context.Context, c *common.Cluster, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetUploadControllerLogsAt", ctx, c, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetUploadControllerLogsAt indicates an expected call of SetUploadControllerLogsAt.
func (mr *MockAPIMockRecorder) SetUploadControllerLogsAt(ctx, c, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUploadControllerLogsAt", reflect.TypeOf((*MockAPI)(nil).SetUploadControllerLogsAt), ctx, c, db)
}

// SetVipsData mocks base method.
func (m *MockAPI) SetVipsData(ctx context.Context, c *common.Cluster, apiVip, ingressVip, apiVipLease, ingressVipLease string, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetVipsData", ctx, c, apiVip, ingressVip, apiVipLease, ingressVipLease, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetVipsData indicates an expected call of SetVipsData.
func (mr *MockAPIMockRecorder) SetVipsData(ctx, c, apiVip, ingressVip, apiVipLease, ingressVipLease, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetVipsData", reflect.TypeOf((*MockAPI)(nil).SetVipsData), ctx, c, apiVip, ingressVip, apiVipLease, ingressVipLease, db)
}

// TransformClusterToDay2 mocks base method.
func (m *MockAPI) TransformClusterToDay2(ctx context.Context, cluster *common.Cluster, db *gorm.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TransformClusterToDay2", ctx, cluster, db)
	ret0, _ := ret[0].(error)
	return ret0
}

// TransformClusterToDay2 indicates an expected call of TransformClusterToDay2.
func (mr *MockAPIMockRecorder) TransformClusterToDay2(ctx, cluster, db interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TransformClusterToDay2", reflect.TypeOf((*MockAPI)(nil).TransformClusterToDay2), ctx, cluster, db)
}

// UpdateAmsSubscriptionID mocks base method.
func (m *MockAPI) UpdateAmsSubscriptionID(ctx context.Context, clusterID, amsSubscriptionID strfmt.UUID) *common.ApiErrorResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAmsSubscriptionID", ctx, clusterID, amsSubscriptionID)
	ret0, _ := ret[0].(*common.ApiErrorResponse)
	return ret0
}

// UpdateAmsSubscriptionID indicates an expected call of UpdateAmsSubscriptionID.
func (mr *MockAPIMockRecorder) UpdateAmsSubscriptionID(ctx, clusterID, amsSubscriptionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAmsSubscriptionID", reflect.TypeOf((*MockAPI)(nil).UpdateAmsSubscriptionID), ctx, clusterID, amsSubscriptionID)
}

// UpdateFinalizingProgress mocks base method.
func (m *MockAPI) UpdateFinalizingProgress(ctx context.Context, db *gorm.DB, clusterID strfmt.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFinalizingProgress", ctx, db, clusterID)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFinalizingProgress indicates an expected call of UpdateFinalizingProgress.
func (mr *MockAPIMockRecorder) UpdateFinalizingProgress(ctx, db, clusterID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFinalizingProgress", reflect.TypeOf((*MockAPI)(nil).UpdateFinalizingProgress), ctx, db, clusterID)
}

// UpdateInstallProgress mocks base method.
func (m *MockAPI) UpdateInstallProgress(ctx context.Context, clusterID strfmt.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateInstallProgress", ctx, clusterID)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateInstallProgress indicates an expected call of UpdateInstallProgress.
func (mr *MockAPIMockRecorder) UpdateInstallProgress(ctx, clusterID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateInstallProgress", reflect.TypeOf((*MockAPI)(nil).UpdateInstallProgress), ctx, clusterID)
}

// UpdateLogsProgress mocks base method.
func (m *MockAPI) UpdateLogsProgress(ctx context.Context, c *common.Cluster, progress string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateLogsProgress", ctx, c, progress)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateLogsProgress indicates an expected call of UpdateLogsProgress.
func (mr *MockAPIMockRecorder) UpdateLogsProgress(ctx, c, progress interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLogsProgress", reflect.TypeOf((*MockAPI)(nil).UpdateLogsProgress), ctx, c, progress)
}

// UploadIngressCert mocks base method.
func (m *MockAPI) UploadIngressCert(c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadIngressCert", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// UploadIngressCert indicates an expected call of UploadIngressCert.
func (mr *MockAPIMockRecorder) UploadIngressCert(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadIngressCert", reflect.TypeOf((*MockAPI)(nil).UploadIngressCert), c)
}

// VerifyClusterUpdatability mocks base method.
func (m *MockAPI) VerifyClusterUpdatability(c *common.Cluster) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyClusterUpdatability", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyClusterUpdatability indicates an expected call of VerifyClusterUpdatability.
func (mr *MockAPIMockRecorder) VerifyClusterUpdatability(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyClusterUpdatability", reflect.TypeOf((*MockAPI)(nil).VerifyClusterUpdatability), c)
}
