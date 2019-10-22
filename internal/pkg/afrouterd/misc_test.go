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

package afrouterd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetIntEnv(t *testing.T) {

	err := os.Setenv("testkey", "123")
	assert.Nil(t, err)

	defer func() { os.Unsetenv("testkey") }()

	v := GetIntEnv("testkey", 0, 1000, 456)
	assert.Equal(t, 123, v)

	v = GetIntEnv("doesnotexist", 0, 1000, 456)
	assert.Equal(t, 456, v)
}

func TestGetIntEnvTooLow(t *testing.T) {

	err := os.Setenv("testkey", "-1")
	assert.Nil(t, err)

	defer func() { os.Unsetenv("testkey") }()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	_ = GetIntEnv("testkey", 0, 1000, 456)
}

func TestGetIntEnvTooHigh(t *testing.T) {

	err := os.Setenv("testkey", "1001")
	assert.Nil(t, err)

	defer func() { os.Unsetenv("testkey") }()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	_ = GetIntEnv("testkey", 0, 1000, 456)
}

func TestGetIntEnvNotInteger(t *testing.T) {

	err := os.Setenv("testkey", "stuff")
	assert.Nil(t, err)

	defer func() { os.Unsetenv("testkey") }()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	_ = GetIntEnv("testkey", 0, 1000, 456)
}

func TestGetStrEnv(t *testing.T) {

	err := os.Setenv("testkey", "abc")
	assert.Nil(t, err)

	defer func() { os.Unsetenv("testkey") }()

	v := GetStrEnv("testkey", "def")
	assert.Equal(t, "abc", v)

	v = GetStrEnv("doesnotexist", "def")
	assert.Equal(t, "def", v)
}
