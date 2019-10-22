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
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha-go/common/log"
	pb "github.com/opencord/voltha-protos/go/afrouter"
	cmn "github.com/opencord/voltha-protos/go/common"
	vpb "github.com/opencord/voltha-protos/go/voltha"
	"golang.org/x/net/context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

func getVolthaPods(cs *kubernetes.Clientset) ([]*volthaPod, error) {
	pods, err := cs.CoreV1().Pods(podNamespace).List(metav1.ListOptions{LabelSelector: podLabelSelector})
	if err != nil {
		return nil, err
	}

	var rwPods []*volthaPod
items:
	for _, v := range pods.Items {
		// only pods that are actually running should be considered
		if v.Status.Phase == v1.PodRunning {
			for _, condition := range v.Status.Conditions {
				if condition.Status != v1.ConditionTrue {
					continue items
				}
			}

			if group, have := v.Labels[podAffinityGroupLabel]; have {
				log.Debugf("Namespace: %s, PodName: %s, PodIP: %s, Host: %s\n", v.Namespace, v.Name, v.Status.PodIP, v.Spec.NodeName)
				rwPods = append(rwPods, &volthaPod{
					name:    v.Name,
					ipAddr:  v.Status.PodIP,
					node:    v.Spec.NodeName,
					devIds:  make(map[string]struct{}),
					backend: afrouterRWClusterName + group,
				})
			} else {
				log.Warnf("Pod %s found matching % without label %", v.Name, podLabelSelector, podAffinityGroupLabel)
			}
		}
	}
	return rwPods, nil
}

func reconcilePodDeviceIds(ctx context.Context, pod *volthaPod, ids map[string]struct{}) {
	ctxTimeout, _ := context.WithTimeout(ctx, time.Second*5)
	conn, err := Connect(ctxTimeout, fmt.Sprintf("%s:%d", pod.ipAddr, podGrpcPort))
	if err != nil {
		log.Debugf("Could not reconcile devices from %s, could not connect: %s", pod.name, err)
		return
	}
	defer conn.Close()

	var idList cmn.IDs
	for k := range ids {
		idList.Items = append(idList.Items, &cmn.ID{Id: k})
	}

	client := vpb.NewVolthaServiceClient(conn)
	_, err = client.ReconcileDevices(ctx, &idList)
	if err != nil {
		log.Errorf("Attempt to reconcile ids on pod %s failed: %s", pod.name, err)
		return
	}
}

func queryPodDeviceIds(ctx context.Context, pod *volthaPod) map[string]struct{} {
	ctxTimeout, _ := context.WithTimeout(ctx, time.Second*5)
	conn, err := Connect(ctxTimeout, fmt.Sprintf("%s:%d", pod.ipAddr, podGrpcPort))
	if err != nil {
		log.Debugf("Could not query devices from %s, could not connect: %s", pod.name, err)
		return nil
	}
	defer conn.Close()

	client := vpb.NewVolthaServiceClient(conn)
	devs, err := client.ListDeviceIds(ctx, &empty.Empty{})
	if err != nil {
		log.Error(err)
		return nil
	}

	var ret = make(map[string]struct{})
	for _, dv := range devs.Items {
		ret[dv.Id] = struct{}{}
	}
	return ret
}

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
