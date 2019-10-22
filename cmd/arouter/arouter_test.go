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

// run the function fp() and return its return value and stdout
func CaptureStdout(fp func() int) (int, string, error) {
	origStdout := os.Stdout

	// log.Cleanup() will call Sync on sys.stdout, and that doesn't
	// work on pipes. Instead of creating a pipe, write the output
	// to a file, then read that file back in.
	f, err := ioutil.TempFile("", "arouter.json")
	if err != nil {
		return 0, "", err
	}

	// Make sure the file is closed and deleted on exit
	defer func() { f.Close(); os.Remove(f.Name()) }()

	// reassign stdout to the file, ensure it will be restored on exit
	os.Stdout = f
	defer func() { os.Stdout = origStdout }()

	status := fp()

	// read back the contents of the tempfile
	_, err = f.Seek(0, 0)
	if err != nil {
		return 0, "", err
	}
	out := make([]byte, 16384)
	numRead, err := f.Read(out)
	if err != nil {
		return 0, "", err
	}

	return status, string(out[:numRead]), nil
}

// Test output of "--version" command
func TestStartupVersionOnly(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	_, err = f.WriteString(config)
	assert.Nil(t, err)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--version", "--config", f.Name()}

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 0, status)

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

func TestStartupMissingConfigFile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	_, err = f.WriteString(config)
	assert.Nil(t, err)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--config", "doesnotexist"}

	// The Voltha logger will write messages to stdout

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 1, status)

	assert.Contains(t, s, "open doesnotexist: no such file or directory")
}

func TestStartupDryRun(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	_, err = f.WriteString(config)
	assert.Nil(t, err)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--dry-run", "--config", f.Name()}

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 0, status)

	assert.Contains(t, s, "Configuration loaded")
}

func TestStartupDryRunGrpcLog(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	_, err = f.WriteString(config)
	assert.Nil(t, err)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--dry-run", "--grpclog", "--config", f.Name()}

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 0, status)

	assert.Contains(t, s, "Configuration loaded")
}

// An unknown command-line option should produce an error
func TestStartupBadCommandLine(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	config, err := MakeConfig()
	assert.Nil(t, err)

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	_, err = f.WriteString(config)
	assert.Nil(t, err)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--dry-run", "--badoption", "--config", f.Name()}

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 1, status)

	assert.Contains(t, s, "Error: Error parsing the command line")
}

// A config file with invalid contents should cause logging output of the error
func TestStartupBadConfigFile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	f, err := ioutil.TempFile("", "arouter.json")
	assert.Nil(t, err)
	_, err = f.WriteString("this is not proper json")
	assert.Nil(t, err)
	f.Close()

	defer func() { os.Remove(f.Name()) }()

	os.Args = []string{os.Args[0], "--dry-run", "--config", f.Name()}

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 1, status)

	assert.Contains(t, s, "invalid character")
}
