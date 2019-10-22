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

package afrouterd

import (
	"github.com/opencord/voltha-go/common/log"
	pb "github.com/opencord/voltha-protos/go/afrouter"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"time"
)

// coreMonitor polls the list of devices from all RW cores, pushes these devices
// into the affinity router, and ensures that all cores in a backend have their devices synced
func CoreMonitor(ctx context.Context, client pb.ConfigurationClient, clientset *kubernetes.Clientset) {
	// map[backend]map[deviceId]struct{}
	deviceOwnership := make(map[string]map[string]struct{})
loop:
	for {
		// get the rw core list from k8s
		rwPods, err := getVolthaPods(clientset)
		if err != nil {
			log.Error(err)
			continue
		}

		// for every pod
		for _, pod := range rwPods {
			// get the devices for this pod's backend
			devices, have := deviceOwnership[pod.backend]
			if !have {
				devices = make(map[string]struct{})
				deviceOwnership[pod.backend] = devices
			}

			coreDevices := queryPodDeviceIds(ctx, pod)

			// handle devices that exist in the core, but we have just learned about
			for deviceId := range coreDevices {
				// if there's a new device
				if _, have := devices[deviceId]; !have {
					// add the device to our local list
					devices[deviceId] = struct{}{}
					// push the device into the affinity router
					setAffinity(ctx, client, deviceId, pod.backend)
				}
			}

			// ensure that the core knows about all devices in its backend
			toSync := make(map[string]struct{})
			for deviceId := range devices {
				// if the pod is missing any devices
				if _, have := coreDevices[deviceId]; !have {
					// we will reconcile them
					toSync[deviceId] = struct{}{}
				}
			}

			if len(toSync) != 0 {
				reconcilePodDeviceIds(ctx, pod, toSync)
			}
		}

		select {
		case <-ctx.Done():
			// if we're done, exit
			break loop
		case <-time.After(10 * time.Second): // wait a while
		}
	}
}
