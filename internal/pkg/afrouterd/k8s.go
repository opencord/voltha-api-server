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
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func K8sClientSet() *kubernetes.Clientset {
	var config *rest.Config
	if k8sApiServer != "" || k8sKubeConfigPath != "" {
		// use combination of URL & local kube-config file
		c, err := clientcmd.BuildConfigFromFlags(k8sApiServer, k8sKubeConfigPath)
		if err != nil {
			panic(err)
		}
		config = c
	} else {
		// use in-cluster config
		c, err := rest.InClusterConfig()
		if err != nil {
			log.Errorf("Unable to load in-cluster config.  Try setting K8S_API_SERVER and K8S_KUBE_CONFIG_PATH?")
			panic(err)
		}
		config = c
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}
