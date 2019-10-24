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
	"flag"
	"fmt"
	"github.com/opencord/voltha-api-server/internal/pkg/afrouterd"
	"github.com/opencord/voltha-lib-go/v2/pkg/version"
	"os"
	"path"

	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	pb "github.com/opencord/voltha-protos/go/afrouter"
	"golang.org/x/net/context"
)

var (
	instanceID         = afrouterd.GetStrEnv("HOSTNAME", "arouterd001")
	afrouterApiAddress = afrouterd.GetStrEnv("AFROUTER_API_ADDRESS", "localhost:55554")
)

type Configuration struct {
	DisplayVersionOnly *bool
}

func startup() int {
	config := &Configuration{}
	cmdParse := flag.NewFlagSet(path.Base(os.Args[0]), flag.ContinueOnError)
	config.DisplayVersionOnly = cmdParse.Bool("version", false, "Print version information and exit")

	if err := cmdParse.Parse(os.Args[1:]); err != nil {
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

	for {
		// Connect to the affinity router
		conn, err := afrouterd.Connect(context.Background(), afrouterApiAddress) // This is a sidecar container so communicating over localhost
		if err != nil {
			panic(err)
		}

		// monitor the connection status, end context if connection is lost
		ctx := afrouterd.ConnectionActiveContext(conn)

		// set up the client
		client := pb.NewConfigurationClient(conn)

		// start the discovery monitor and core monitor
		// these two processes do the majority of the work

		log.Info("Starting discovery monitoring")
		doneCh, _ := afrouterd.StartDiscoveryMonitor(ctx, client)

		log.Info("Starting core monitoring")
		afrouterd.CoreMonitor(ctx, client, clientset)

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
