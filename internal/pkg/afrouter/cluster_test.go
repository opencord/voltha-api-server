/*
 * Copyright 2019-present Open Networking Foundation

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

func makeBackendClusterConfig(numBackends int, numConnections int) *BackendClusterConfig {

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

func TestNewBackendCluster(t *testing.T) {
	backends := makeBackendClusterConfig(1, 1)
	cluster, err := newBackendCluster(backends)
	assert.NotNil(t, cluster)
	assert.Nil(t, err)
}

func TestNewBackendClusterTwoOne(t *testing.T) {
	backends := makeBackendClusterConfig(2, 1)
	cluster, err := newBackendCluster(backends)
	assert.NotNil(t, cluster)
	assert.Nil(t, err)
}

func TestNewBackendClusterNoName(t *testing.T) {
	backends := makeBackendClusterConfig(1, 1)
	backends.Name = ""
	backend, err := newBackendCluster(backends)
	assert.Nil(t, backend)
	assert.NotNil(t, err, "A backend cluster must have a name")
}

func TestNewBackendClusterBackendsNoName(t *testing.T) {
	backends := makeBackendClusterConfig(1, 1)
	backends.Backends[0].Name = ""
	backend, err := newBackendCluster(backends)
	assert.Nil(t, backend)
	assert.NotNil(t, err, "A backend cluster must have a name")
}

func TestNewBackendClusterBackendsTwoNoName(t *testing.T) {
	backends := makeBackendClusterConfig(2, 1)
	//backends.Backends[0].Name = ""
	backends.Backends[1].Name = ""
	backend, err := newBackendCluster(backends)
	assert.Nil(t, backend)
	assert.NotNil(t, err, "A backend cluster must have a name")
}
