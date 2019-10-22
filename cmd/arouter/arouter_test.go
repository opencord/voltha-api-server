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

// This file implements an exit handler that tries to shut down all the
// running servers before finally exiting. There are 2 triggers to this
// clean exit thread: signals and an exit channel.

package main

import (
	"fmt"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

// Generate a configuration, ensure any ports are randomly chosen free ports
func MakeConfig() (string, error) {
	freePort, err := freeport.GetFreePort()
	if err != nil {
		return "", err
	}

	config := fmt.Sprintf(`{
	"api": {
		"_comment": "If this isn't defined then no api is available for dynamic configuration and queries",
		"address": "localhost",
		"port": %d
	}
 }`, freePort)

	return config, nil
}

// run the function fp() and return its stdout
func CaptureStdout(fp func()) string {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	fp()

	w.Close()
	out, _ := ioutil.ReadAll(r)

	return string(out)
}

// run the function fp() and return its stderr
func CaptureLog(fp func()) string {
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	fp()

	w.Close()
	out, _ := ioutil.ReadAll(r)

	return string(out)
}

// Test output of "--version" command
func TestMainVersionOnly(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	f.WriteString(config)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--version", "--config", f.Name()}

	s := CaptureStdout(main)

	expected := `VOLTHA API Server (afrouter)
  Version:      unknown-version
  GoVersion:    unknown-goversion
  VCS Ref:      unknown-vcsref
  VCS Dirty:    unknown-vcsdirty
  Built:        unknown-buildtime
  OS/Arch:      unknown-os/unknown-arch

`
	assert.Equal(t, expected, s)
}

func TestMainMissingConfigFile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	f.WriteString(config)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--config", "doesnotexist"}

	// The Voltha logger will write messages to stdout

	s := CaptureStdout(main)

	assert.Contains(t, s, "open doesnotexist: no such file or directory")
}

func TestMainDryRun(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	f.WriteString(config)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--dry-run", "--config", f.Name()}

	s := CaptureStdout(main)

	assert.Contains(t, s, "Configuration loaded")
}
