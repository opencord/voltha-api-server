/*
 * Portions copyright 2019-present Open Networking Foundation
 * Original copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the"github.com/stretchr/testify/assert" "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package afrouter

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	voltha_pb "github.com/opencord/voltha-protos/v2/go/voltha"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

// MockContext, always returns the specified Metadata

type MockContext struct {
	metadata metadata.MD
}

func (mc MockContext) Deadline() (deadline time.Time, ok bool) { return time.Now(), true }
func (mc MockContext) Done() <-chan struct{}                   { return nil }
func (mc MockContext) Err() error                              { return nil }
func (mc MockContext) Value(key interface{}) interface{} {
	return mc.metadata
}

// MockServerStream, always returns a Context that returns the specified metadata

type MockServerStream struct {
	metadata metadata.MD
}

func (ms MockServerStream) SetHeader(_ metadata.MD) error  { return nil }
func (ms MockServerStream) SendHeader(_ metadata.MD) error { return nil }
func (ms MockServerStream) SetTrailer(_ metadata.MD)       {}
func (ms MockServerStream) Context() context.Context       { return MockContext(ms) }
func (ms MockServerStream) SendMsg(_ interface{}) error    { return nil }
func (ms MockServerStream) RecvMsg(_ interface{}) error    { return nil }

// Build an method router configuration
func MakeBindingTestConfig(numBackends int, numConnections int) (*RouteConfig, *RouterConfig) {
	var backends []BackendConfig
	for backendIndex := 0; backendIndex < numBackends; backendIndex++ {
		var connections []ConnectionConfig
		for connectionIndex := 0; connectionIndex < numConnections; connectionIndex++ {
			connectionConfig := ConnectionConfig{
				Name: fmt.Sprintf("rw_vcore%d%d", backendIndex, connectionIndex+1),
				Addr: "foo",
				Port: "123",
			}
			connections = append(connections, connectionConfig)
		}

		backendConfig := BackendConfig{
			Name:        fmt.Sprintf("rw_vcore%d", backendIndex),
			Type:        BackendSingleServer,
			Connections: connections,
		}

		backends = append(backends, backendConfig)
	}

	backendClusterConfig := BackendClusterConfig{
		Name:     "vcore",
		Backends: backends,
	}

	bindingConfig := BindingConfig{
		Type:        "header",
		Field:       "voltha_backend_name",
		Method:      "Subscribe",
		Association: AssociationRoundRobin,
	}

	routeConfig := RouteConfig{
		Name:             "dev_manager_ofagent",
		Type:             RouteTypeRpcAffinityMessage,
		Association:      AssociationRoundRobin,
		BackendCluster:   "vcore",
		backendCluster:   &backendClusterConfig,
		Binding:          bindingConfig,
		RouteField:       "id",
		Methods:          []string{"CreateDevice", "EnableDevice"},
		NbBindingMethods: []string{"CreateDevice"},
	}

	routerConfig := RouterConfig{
		Name:         "vcore",
		ProtoService: "VolthaService",
		ProtoPackage: "voltha",
		Routes:       []RouteConfig{routeConfig},
		ProtoFile:    TEST_PROTOFILE,
	}
	return &routeConfig, &routerConfig
}

// Route() requires an open connection, so pretend we have one.
func PretendBindingOpenConnection(router Router, clusterName string, backendIndex int, connectionName string) {
	cluster := router.FindBackendCluster(clusterName)

	// Route Method expects an open connection
	conn := cluster.backends[backendIndex].connections[connectionName]
	cluster.backends[backendIndex].openConns[conn] = &grpc.ClientConn{}
}

// Common setup to run before each unit test
func BindingTestSetup() {
	// reset globals that need to be clean for each unit test

	clusters = make(map[string]*cluster)
	allRouters = make(map[string]Router)
}

// Test creation of a new AffinityRouter, and the Service(), Name(), FindBackendCluster(), and
// methods.
func TestBindingRouterInit(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.Nil(t, err)

	assert.Equal(t, router.Service(), "VolthaService")
	assert.Equal(t, router.Name(), "dev_manager_ofagent")

	cluster, err := router.BackendCluster("EnableDevice", NoMeta)
	assert.Equal(t, cluster, clusters["vcore"])
	assert.Nil(t, err)

	assert.Equal(t, router.FindBackendCluster("vcore"), clusters["vcore"])
}

// Passing no ProtoPackage should return an error
func TestBindingRouterNoProtoPackage(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	routerConfig.ProtoPackage = ""

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.EqualError(t, err, "Failed to create a new router 'dev_manager_ofagent'")
}

// Passing no ProtoService should return an error
func TestBindingRouterNoProtoService(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	routerConfig.ProtoService = ""

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.EqualError(t, err, "Failed to create a new router 'dev_manager_ofagent'")
}

// Passing no ProtoService should return an error
func TestBindingRouterNoAssociation(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	routeConfig.Binding.Association = AssociationUndefined

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.EqualError(t, err, "Failed to create a new router 'dev_manager_ofagent'")
}

// Passing type other than "header" should return an error
func TestBindingRouterInvalidType(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	routeConfig.Binding.Type = "wrong"

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.EqualError(t, err, "Failed to create a new router 'dev_manager_ofagent'")
}

// Passing no Method should return an error
func TestBindingNoMethod(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	routeConfig.Binding.Method = ""

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.EqualError(t, err, "Failed to create a new router 'dev_manager_ofagent'")
}

// Passing no Field should return an error
func TestBindingNoField(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	routeConfig.Binding.Method = ""

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.EqualError(t, err, "Failed to create a new router 'dev_manager_ofagent'")
}

func TestBindingRouterGetMetaKeyVal(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.Nil(t, err)

	ms := MockServerStream{}
	ms.metadata = make(map[string][]string)
	ms.metadata["voltha_backend_name"] = []string{"some_backend"}

	k, v, err := router.GetMetaKeyVal(ms)

	assert.Nil(t, err)
	assert.Equal(t, "voltha_backend_name", k)
	assert.Equal(t, "some_backend", v)
}

// If metadata doesn't exist, return empty strings
func TestBindingRouterGetMetaKeyValEmptyMetadata(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.Nil(t, err)

	ms := MockServerStream{}
	ms.metadata = nil

	k, v, err := router.GetMetaKeyVal(ms)

	assert.Nil(t, err)
	assert.Equal(t, "", k)
	assert.Equal(t, "", v)
}

func TestBindingRoute(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.Nil(t, err)

	ms := MockServerStream{}
	ms.metadata = nil

	subscribeMessage := &voltha_pb.OfAgentSubscriber{OfagentId: "1234", VolthaId: "5678"}

	subscribeData, err := proto.Marshal(subscribeMessage)
	assert.Nil(t, err)

	sel := &requestFrame{payload: subscribeData,
		err:        nil,
		metaKey:    "voltha_backend_name",
		metaVal:    "",
		methodInfo: newMethodDetails("/voltha.VolthaService/Subscribe")}

	PretendBindingOpenConnection(router, "vcore", 0, "rw_vcore01")

	backend, connection := router.Route(sel)
	assert.NotNil(t, backend)
	assert.Nil(t, connection)

	// now route it again with a metaVal, and we should find the existing binding

	sel = &requestFrame{payload: subscribeData,
		err:        nil,
		metaKey:    "voltha_backend_name",
		metaVal:    "rw_vcore0",
		methodInfo: newMethodDetails("/voltha.VolthaService/Subscribe")}

	backend, connection = router.Route(sel)
	assert.NotNil(t, backend)
	assert.Nil(t, connection)
}

// Only "Subscribe" is a valid method
func TestBindingRouteWrongMethod(t *testing.T) {
	BindingTestSetup()

	routeConfig, routerConfig := MakeBindingTestConfig(1, 1)

	router, err := newBindingRouter(routerConfig, routeConfig)

	assert.NotNil(t, router)
	assert.Nil(t, err)

	ms := MockServerStream{}
	ms.metadata = nil

	subscribeMessage := &voltha_pb.OfAgentSubscriber{OfagentId: "1234", VolthaId: "5678"}

	subscribeData, err := proto.Marshal(subscribeMessage)
	assert.Nil(t, err)

	sel := &requestFrame{payload: subscribeData,
		err:        nil,
		metaKey:    "voltha_backend_name",
		metaVal:    "",
		methodInfo: newMethodDetails("/voltha.VolthaService/EnableDevice")}

	PretendBindingOpenConnection(router, "vcore", 0, "rw_vcore01")

	backend, connection := router.Route(sel)
	assert.Nil(t, backend)
	assert.Nil(t, connection)
}
