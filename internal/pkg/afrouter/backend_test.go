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

// Unit Test Backend manager that handles redundant connections per backend

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func makeBackend() *BackendConfig {
	var connections []ConnectionConfig
	connectionConfig := ConnectionConfig{
		Name: fmt.Sprintf("ro_vcore%d%d", 0, 1),
		Addr: "foo",
		Port: "123",
	}
	connections = append(connections, connectionConfig)

	backendConfig := BackendConfig{
		Name:        fmt.Sprintf("ro_vcore%d", 0),
		Type:        BackendSingleServer,
		Connections: connections,
	}
	return &backendConfig
}

func makeBackendCluster(numBackends int, numConnections int) *BackendClusterConfig {

	var backends []BackendConfig
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
	return &backendClusterConfig
}

func TestBackend(t *testing.T) {

	backend, err := newBackend(makeBackend(), "Cluster")
	assert.NotNil(t, backend)
	assert.Nil(t, err)
}

func TestBackendInvalidType(t *testing.T) {
	conf := makeBackend()
	conf.Type = BackendUndefined
	backend, err := newBackend(conf, "Cluster")
	assert.Nil(t, backend)
	assert.NotNil(t, err, "Backend Invalid Type Undefined")
}

func TestBackendAssociationLocationMissing(t *testing.T) {
	conf := makeBackend()
	conf.Type = BackendActiveActive
	conf.Association.Location = AssociationLocationUndefined
	backend, err := newBackend(conf, "Cluster")
	assert.Nil(t, backend)
	assert.NotNil(t, err, "An association location must be provided if the backend "+
		"type is active/active for backend in cluster ")
}

func TestBackendAssociationFieldMissing(t *testing.T) {
	conf := makeBackend()
	conf.Association.Location = AssociationLocationProtobuf
	conf.Association.Field = ""
	backend, err := newBackend(conf, "Cluster")
	assert.Nil(t, backend)
	assert.NotNil(t, err, "An association field must be provided if the backend "+
		"type is active/active and the location is set to protobuf "+
		"for backend in cluster")
}

func TestBackendAssociationStrategyMissing(t *testing.T) {
	conf := makeBackend()
	conf.Association.Strategy = AssociationStrategyUndefined
	conf.Type = BackendActiveActive
	backend, err := newBackend(conf, "Cluster")
	assert.Nil(t, backend)
	assert.NotNil(t, err, "An association strategy must be provided if the backend "+
		"type is active/active")
}

func TestBackendAssociationKeyMissing(t *testing.T) {
	conf := makeBackend()
	conf.Association.Key = ""
	conf.Association.Location = AssociationLocationHeader
	backend, err := newBackend(conf, "Cluster")
	assert.Nil(t, backend)
	assert.NotNil(t, err, "An association key must be provided if the backend "+
		"type is active/active and the location is set to header "+
		"for backend in cluster")
}
func TestBackendWithMoreConnections(t *testing.T) {
	conf := makeBackendCluster(1, 4)
	conf.Backends[0].Association.Location = AssociationLocationHeader
	conf.Backends[0].Association.Key = "BackendKey"
	backend, err := newBackend(&conf.Backends[0], "Cluster")

	assert.Nil(t, backend)
	assert.NotNil(t, err, "Only one connection must be specified if the association "+
		"strategy is not set to 'active_active'")

}
func TestBackendNoConnections(t *testing.T) {
	confZero := makeBackendCluster(1, 0)
	confZero.Backends[0].Association.Location = AssociationLocationHeader
	confZero.Backends[0].Association.Key = "BackendKey"
	backendNoConn, err := newBackend(&confZero.Backends[0], "Cluster")

	assert.Nil(t, backendNoConn)
	assert.NotNil(t, err, "A connection must have a name for backend in cluster")
}
