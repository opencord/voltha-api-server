/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package afrouter

import (
	"fmt"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"testing"
)

func makeConfig(numBackends int, numConnections int) Configuration {

	var backends []BackendConfig
	var slbackendClusterConfig []BackendClusterConfig
	var slRouterConfig []RouterConfig

	for backendIndex := 0; backendIndex < numBackends; backendIndex++ {
		var connections []ConnectionConfig
		for connectionIndex := 0; connectionIndex < numConnections; connectionIndex++ {
			connectionConfig := ConnectionConfig{
				Name: fmt.Sprintf("ro_vcore%d%d", backendIndex, connectionIndex+1),
				Addr: "foo",
				Port: "123",
			}
			connections = append(connections, connectionConfig)
		}

		backendConfig := BackendConfig{
			Name:        fmt.Sprintf("ro_vcore%d", backendIndex),
			Type:        BackendSingleServer,
			Connections: connections,
		}

		backends = append(backends, backendConfig)
	}

	backendClusterConfig := BackendClusterConfig{
		Name:     "ro_vcore",
		Backends: backends,
	}

	slbackendClusterConfig = append(slbackendClusterConfig, backendClusterConfig)

	routeConfig := RouteConfig{
		Name:           "read_only",
		Type:           RouteTypeRoundRobin,
		Association:    AssociationRoundRobin,
		BackendCluster: "ro_vcore",
		backendCluster: &backendClusterConfig,
		Methods:        []string{"ListDevicePorts"},
	}

	routerConfig := RouterConfig{
		Name:         "vcore",
		ProtoService: "VolthaService",
		ProtoPackage: "voltha",
		Routes:       []RouteConfig{routeConfig},
		ProtoFile:    TEST_PROTOFILE,
	}

	slRouterConfig = append(slRouterConfig, routerConfig)

	port1, _ := freeport.GetFreePort()
	apiConfig := ApiConfig{
		Addr: "127.0.0.1",
		Port: uint(port1),
	}

	cfile := "config_file"
	loglevel := 0
	glog := false

	conf := Configuration{
		InstanceID:      "1",
		ConfigFile:      &cfile,
		LogLevel:        &loglevel,
		GrpcLog:         &glog,
		BackendClusters: slbackendClusterConfig,
		Routers:         slRouterConfig,
		Api:             apiConfig,
	}
	return conf
}

func makeProxy(numBackends int, numConnections int) (*ArouterProxy, error) {

	conf := makeConfig(3, 2)
	arouter, err := NewArouterProxy(&conf)
	return arouter, err
}

func TestNewApi(t *testing.T) {

	port, _ := freeport.GetFreePort()

	apiConfig := ApiConfig{
		Addr: "127.0.0.1",
		Port: uint(port),
	}
	arouter, err := makeProxy(3, 2)
	assert.NotNil(t, arouter, "Error the proxy confiiguration fails!")
	assert.Nil(t, err)

	ar, err := newApi(&apiConfig, arouter)

	assert.NotNil(t, ar, "NewApi Fails!")
	assert.Nil(t, err)
}

func TestNewApiWrongSocket(t *testing.T) {
	port, _ := freeport.GetFreePort()
	apiConfig := ApiConfig{
		Addr: "256.300.1.83",
		Port: uint(port),
	}
	arouter, err := makeProxy(3, 2)
	assert.NotNil(t, arouter, "Error the proxy configuration fails!")
	assert.Nil(t, err)
	ar, err := newApi(&apiConfig, arouter)

	assert.Nil(t, ar, "Accepted a wrong IP")
	assert.NotNil(t, err)

	apiConfig = ApiConfig{
		Addr: "127.0.0.1",
		Port: 68000,
	}
	ar, err = newApi(&apiConfig, arouter)
	assert.Nil(t, ar, "Accepted a wrong port")
	assert.NotNil(t, err)
}

func TestGetBackend(t *testing.T) {
	port, _ := freeport.GetFreePort()
	apiConfig := ApiConfig{
		Addr: "127.0.0.1",
		Port: uint(port),
	}
	arouter, _ := makeProxy(3, 2)

	ar, _ := newApi(&apiConfig, arouter)

	backends := makeBackendClusterConfig(5, 1)
	cluster, err := newBackendCluster(backends)
	assert.NotNil(t, cluster)
	assert.Nil(t, err)

	for backendIndex := 0; backendIndex < 5; backendIndex++ {
		backend, errb := ar.getBackend(cluster, fmt.Sprintf("ro_vcore%d", backendIndex))
		assert.NotNil(t, backend)
		assert.Nil(t, errb)
	}

	backend, errb := ar.getBackend(cluster, "ro_vcore5")
	assert.Nil(t, backend)
	assert.NotNil(t, errb)
}

func TestGetConnection(t *testing.T) {
	port, _ := freeport.GetFreePort()
	apiConfig := ApiConfig{
		Addr: "127.0.0.1",
		Port: uint(port),
	}
	arouter, _ := makeProxy(3, 2)

	ar, _ := newApi(&apiConfig, arouter)

	backend, err := newBackend(makeBackend(), "Cluster")
	assert.NotNil(t, backend)
	assert.Nil(t, err)
	cc, _ := ar.getConnection(backend, "ro_vcore01")
	assert.NotNil(t, cc, "Error connection name not found")

	cc2, err2 := ar.getConnection(backend, "wrongConnectionName")
	assert.Nil(t, cc2)
	assert.NotNil(t, err2)
}
