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
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

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

	os.Args = []string{os.Args[0], "--version"}

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 0, status)

	expected := `VOLTHA API Server (afrouterd)
  Version:      unknown-version
  GoVersion:    unknown-goversion
  VCS Ref:      unknown-vcsref
  VCS Dirty:    unknown-vcsdirty
  Built:        unknown-buildtime
  OS/Arch:      unknown-os/unknown-arch

`
	assert.Equal(t, expected, s)
}

// An unknown command-line option should produce an error
func TestStartupBadCommandLine(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{os.Args[0], "--badoption"}

	status, s, err := CaptureStdout(startup)
	assert.Nil(t, err)

	assert.Equal(t, 1, status)

	assert.Contains(t, s, "Error: flag provided but not defined: -badoption")
}
