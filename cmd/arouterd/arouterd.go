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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/opencord/voltha-api-server/internal/pkg/afrouterd"
	"github.com/opencord/voltha-lib-go/v2/pkg/probe"
	"github.com/opencord/voltha-lib-go/v2/pkg/version"
	"os"
	"path"

	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	pb "github.com/opencord/voltha-protos/go/afrouter"
)

var (
	instanceID         = afrouterd.GetStrEnv("HOSTNAME", "arouterd001")
	afrouterApiAddress = afrouterd.GetStrEnv("AFROUTER_API_ADDRESS", "localhost:55554")
)

type Configuration struct {
	DisplayVersionOnly *bool
	ProbeHost          *string
	ProbePort          *int
}

func startup() int {
	config := &Configuration{}
	cmdParse := flag.NewFlagSet(path.Base(os.Args[0]), flag.ContinueOnError)
	config.DisplayVersionOnly = cmdParse.Bool("version", false, "Print version information and exit")
	config.ProbeHost = cmdParse.String("probeHost", "", "probe host info")
	config.ProbePort = cmdParse.Int("probePort", 8081, "probe port info")

	if err := cmdParse.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	if *config.DisplayVersionOnly {
		fmt.Println("VOLTHA API Server (afrouterd)")
		fmt.Println(version.VersionInfo.String("  "))
		return 0
	}

	// Set up logging
	if _, err := log.SetDefaultLogger(log.JSON, 0, log.Fields{"instanceId": instanceID}); err != nil {
		log.With(log.Fields{"error": err}).Fatal("Cannot setup logging")
	}

	// Set up kubernetes api
	clientset := afrouterd.K8sClientSet()

	/*
	 * Create and start the liveness and readiness container management probes. This
	 * is done in the main function so just in case the main starts multiple other
	 * objects there can be a single probe end point for the process.
	 */
	ctx, cancel := context.WithCancel(context.Background())

	p := &probe.Probe{}
	go p.ListenAndServe(fmt.Sprintf("%s:%d", *config.ProbeHost, *config.ProbePort))
	//go p.ListenAndServe(*config.ProbeHost, *config.ProbePort)

	// Add the probe to the context to pass to all the services started
	probeCtx := context.WithValue(ctx, probe.ProbeContextKey, p)

	// If the context has a probe then fetch it and register our services
	if value := ctx.Value(probe.ProbeContextKey); value != nil {
		if _, ok := value.(*probe.Probe); ok {
			p = value.(*probe.Probe)
			p.RegisterService(
				"affinity-router",
				"message-bus",
			)
		}
	}

	for {
		// Connect to the affinity router
		conn, err := afrouterd.Connect(probeCtx, afrouterApiAddress) // This is a sidecar container so communicating over localhost
		if err != nil {
			panic(err)
		}
		if p != nil {
			p.UpdateStatus("affinity-router", probe.ServiceStatusRunning)
		}

		// monitor the connection status, end context if connection is lost
		afrouterd.ConnectionActiveContext(conn, probeCtx, cancel)

		// set up the client
		client := pb.NewConfigurationClient(conn)

		// start the discovery monitor and core monitor
		// these two processes do the majority of the work

		log.Info("Starting discovery monitoring")
		doneCh, _ := afrouterd.StartDiscoveryMonitor(probeCtx, client)

		log.Info("Starting core monitoring")
		afrouterd.CoreMonitor(probeCtx, client, clientset)

		//ensure the discovery monitor to quit
		<-doneCh

		conn.Close()
	}
}

func main() {
	status := startup()
	if status != 0 {
		os.Exit(status)
	}
}
