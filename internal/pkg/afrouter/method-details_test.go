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

import (
	"reflect"
	"testing"
)

func Test_newMethodDetails(t *testing.T) {
	type args struct {
		fullMethodName string
	}
	tests := []struct {
		name string
		args args
		want methodDetails
	}{
		// TODO: Add test cases.
		{"newMethodDetails-1", args{fullMethodName: "/voltha.VolthaService/EnableDevice"},
			methodDetails{all: "/voltha.VolthaService/EnableDevice", pkg: "voltha", service: "VolthaService", method: "EnableDevice"}},
		{"newMethodDetails-2", args{fullMethodName: "/voltha.VolthaService/DisableDevice"},
			methodDetails{all: "/voltha.VolthaService/DisableDevice", pkg: "voltha", service: "VolthaService", method: "DisableDevice"}},
		{"newMethodDetails-3", args{fullMethodName: "/voltha.VolthaService/ListDevicePorts"},
			methodDetails{all: "/voltha.VolthaService/ListDevicePorts", pkg: "voltha", service: "VolthaService", method: "ListDevicePorts"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newMethodDetails(tt.args.fullMethodName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMethodDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
