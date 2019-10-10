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
	"github.com/stretchr/testify/assert"
	"testing"
)

func MakeServerTestConfig() *ServerConfig {
	var routerPackage []RouterPackage

	routerPackageConfig := RouterPackage{
		Router:  `json:"router"`,
		Package: `json:"package"`,
	}
	routerPackage = append(routerPackage, routerPackageConfig)

	serverConfig := ServerConfig{
		Name:    "grpc_command",
		Port:    55555,
		Addr:    "127.0.0.1",
		Type:    "grpc",
		Routers: routerPackage,
		routers: make(map[string]*RouterConfig),
	}
	return &serverConfig

}

// Test creation of a new Server
func TestServerInit(t *testing.T) {

	serverConfig := MakeServerTestConfig()

	serv, err := newServer(serverConfig)

	assert.NotNil(t, serv)
	assert.Nil(t, err)

}

// Test creation of a new Server, error in Addr
func TestServerInitWrongAddr(t *testing.T) {

	serverConfig := MakeServerTestConfig()
	serverConfig.Addr = "127.300.1.1"

	serv, err := newServer(serverConfig)

	assert.Nil(t, serv)
	assert.NotNil(t, err)
}

// Test creation of a new Server, error in Port
func TestServerInitWrongPort(t *testing.T) {

	serverConfig := MakeServerTestConfig()
	serverConfig.Port = 23

	serv, err := newServer(serverConfig)

	assert.Nil(t, serv)
	assert.NotNil(t, err)
}

// Test creation of a new Server, error in Name
func TestServerInitNoName(t *testing.T) {

	serverConfig := MakeServerTestConfig()
	serverConfig.Name = ""

	serv, err := newServer(serverConfig)

	assert.Nil(t, serv)
	assert.NotNil(t, err)
}

// Test creation of a new Server, error in Type
func TestServerInitWrongType(t *testing.T) {

	serverConfig := MakeServerTestConfig()
	serverConfig.Type = "xxx"

	serv, err := newServer(serverConfig)

	assert.Nil(t, serv)
	assert.NotNil(t, err)
}

// Test creation of a new Server, error in Router
func TestServerInitNoRouter(t *testing.T) {

	serverConfig := MakeServerTestConfig()
	serverConfig.routers = nil

	serv, err := newServer(serverConfig)

	assert.Nil(t, serv)
	assert.NotNil(t, err)
}

// Test creation of a new Server
func TestServerInitHandler(t *testing.T) {

	serverConfig := MakeServerTestConfig()
	serverConfig.Port = 55556

	serv, err := newServer(serverConfig)

	assert.NotNil(t, serv)
	assert.Nil(t, err)

}
