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
	"flag"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func MakeConfigTestConfig() *Configuration {

	var confFilePath string

	config := Configuration{}

	cmdParse := flag.NewFlagSet(path.Base(os.Args[0]), flag.ContinueOnError)
	confFilePath = fmt.Sprintf("%s/src/github.com/opencord/voltha-api-server/arouter.json", os.Getenv("GOPATH"))
	config.ConfigFile = cmdParse.String("config", confFilePath, "The configuration file for the affinity router")

	return &config

}

// Test Config
func TestConfigConfig(t *testing.T) {

	var err error

	configConf := MakeConfigTestConfig()
	assert.NotNil(t, configConf)

	err = configConf.LoadConfig()
	assert.Nil(t, err)

}

// Test Config with wrong config file
func TestConfigNoFile(t *testing.T) {

	var err error
	var confWrongFilePath string

	configConf := MakeConfigTestConfig()
	assert.NotNil(t, configConf)

	confWrongFilePath = fmt.Sprintf("%s/src/github.com/opencord/voltha-api-server/xxx.json", os.Getenv("GOPATH"))
	configConf.ConfigFile = &confWrongFilePath

	err = configConf.LoadConfig()
	assert.NotNil(t, err)

}
