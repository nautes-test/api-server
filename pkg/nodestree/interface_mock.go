// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/nodestree/interface.go

// Package nodestree is a generated GoMock package.
package nodestree

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockNodesTree is a mock of NodesTree interface.
type MockNodesTree struct {
	ctrl     *gomock.Controller
	recorder *MockNodesTreeMockRecorder
}

// MockNodesTreeMockRecorder is the mock recorder for MockNodesTree.
type MockNodesTreeMockRecorder struct {
	mock *MockNodesTree
}

// NewMockNodesTree creates a new mock instance.
func NewMockNodesTree(ctrl *gomock.Controller) *MockNodesTree {
	mock := &MockNodesTree{ctrl: ctrl}
	mock.recorder = &MockNodesTreeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNodesTree) EXPECT() *MockNodesTreeMockRecorder {
	return m.recorder
}

// AppendOperators mocks base method.
func (m *MockNodesTree) AppendOperators(operator NodesOperator) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AppendOperators", operator)
}

// AppendOperators indicates an expected call of AppendOperators.
func (mr *MockNodesTreeMockRecorder) AppendOperators(operator interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendOperators", reflect.TypeOf((*MockNodesTree)(nil).AppendOperators), operator)
}

// Compare mocks base method.
func (m *MockNodesTree) Compare(options CompareOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Compare", options)
	ret0, _ := ret[0].(error)
	return ret0
}

// Compare indicates an expected call of Compare.
func (mr *MockNodesTreeMockRecorder) Compare(options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Compare", reflect.TypeOf((*MockNodesTree)(nil).Compare), options)
}

// GetNode mocks base method.
func (m *MockNodesTree) GetNode(nodes *Node, kind, name string) *Node {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNode", nodes, kind, name)
	ret0, _ := ret[0].(*Node)
	return ret0
}

// GetNode indicates an expected call of GetNode.
func (mr *MockNodesTreeMockRecorder) GetNode(nodes, kind, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNode", reflect.TypeOf((*MockNodesTree)(nil).GetNode), nodes, kind, name)
}

// InsertNodes mocks base method.
func (m *MockNodesTree) InsertNodes(nodes, resource *Node) (*Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertNodes", nodes, resource)
	ret0, _ := ret[0].(*Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertNodes indicates an expected call of InsertNodes.
func (mr *MockNodesTreeMockRecorder) InsertNodes(nodes, resource interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertNodes", reflect.TypeOf((*MockNodesTree)(nil).InsertNodes), nodes, resource)
}

// Load mocks base method.
func (m *MockNodesTree) Load(path string) (Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Load", path)
	ret0, _ := ret[0].(Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Load indicates an expected call of Load.
func (mr *MockNodesTreeMockRecorder) Load(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Load", reflect.TypeOf((*MockNodesTree)(nil).Load), path)
}

// RemoveNode mocks base method.
func (m *MockNodesTree) RemoveNode(nodes, node *Node) (*Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveNode", nodes, node)
	ret0, _ := ret[0].(*Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveNode indicates an expected call of RemoveNode.
func (mr *MockNodesTreeMockRecorder) RemoveNode(nodes, node interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveNode", reflect.TypeOf((*MockNodesTree)(nil).RemoveNode), nodes, node)
}

// MockNodesOperator is a mock of NodesOperator interface.
type MockNodesOperator struct {
	ctrl     *gomock.Controller
	recorder *MockNodesOperatorMockRecorder
}

// MockNodesOperatorMockRecorder is the mock recorder for MockNodesOperator.
type MockNodesOperatorMockRecorder struct {
	mock *MockNodesOperator
}

// NewMockNodesOperator creates a new mock instance.
func NewMockNodesOperator(ctrl *gomock.Controller) *MockNodesOperator {
	mock := &MockNodesOperator{ctrl: ctrl}
	mock.recorder = &MockNodesOperatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNodesOperator) EXPECT() *MockNodesOperatorMockRecorder {
	return m.recorder
}

// CheckReference mocks base method.
func (m *MockNodesOperator) CheckReference(options CompareOptions, node *Node, k8sClient client.Client) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckReference", options, node, k8sClient)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckReference indicates an expected call of CheckReference.
func (mr *MockNodesOperatorMockRecorder) CheckReference(options, node, k8sClient interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckReference", reflect.TypeOf((*MockNodesOperator)(nil).CheckReference), options, node, k8sClient)
}

// CreateNode mocks base method.
func (m *MockNodesOperator) CreateNode(path string, data interface{}) (*Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateNode", path, data)
	ret0, _ := ret[0].(*Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateNode indicates an expected call of CreateNode.
func (mr *MockNodesOperatorMockRecorder) CreateNode(path, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNode", reflect.TypeOf((*MockNodesOperator)(nil).CreateNode), path, data)
}

// CreateResource mocks base method.
func (m *MockNodesOperator) CreateResource(kind string) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateResource", kind)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// CreateResource indicates an expected call of CreateResource.
func (mr *MockNodesOperatorMockRecorder) CreateResource(kind interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResource", reflect.TypeOf((*MockNodesOperator)(nil).CreateResource), kind)
}

// UpdateNode mocks base method.
func (m *MockNodesOperator) UpdateNode(node *Node, data interface{}) (*Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNode", node, data)
	ret0, _ := ret[0].(*Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateNode indicates an expected call of UpdateNode.
func (mr *MockNodesOperatorMockRecorder) UpdateNode(node, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNode", reflect.TypeOf((*MockNodesOperator)(nil).UpdateNode), node, data)
}