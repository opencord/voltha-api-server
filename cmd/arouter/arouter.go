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
	"fmt"
	"github.com/opencord/voltha-api-server/internal/pkg/afrouter"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-lib-go/v2/pkg/version"
	_ "github.com/opencord/voltha-protos/v2"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
	"os"
)

// startup arouter, return exit status as an integer
func startup() int {

	conf, err := afrouter.ParseCmd()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	// Setup logging
	if _, err := log.SetDefaultLogger(log.JSON, *conf.LogLevel, log.Fields{"instanceId": conf.InstanceID}); err != nil {
		log.With(log.Fields{"error": err}).Fatal("Cannot setup logging")
		return 1
	}

	defer func() {
		err := log.CleanUp()
		if err != nil {
			// Let's not use the logger to print the error message, since the
			// logger could be in a bad state.
			fmt.Fprintf(os.Stderr, "Failed to cleanup logger: %v", err)
		}
	}()

	if *conf.DisplayVersionOnly {
		fmt.Println("VOLTHA API Server (afrouter)")
		fmt.Println(version.VersionInfo.String("  "))
		return 0
	}

	// Parse the config file
	err = conf.LoadConfig()
	if err != nil {
		log.Error(err)
		return 1
	}
	log.With(log.Fields{"config": *conf}).Debug("Configuration loaded")

	// Enable grpc logging
	if *conf.GrpcLog {
		grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stderr, ioutil.Discard, ioutil.Discard))
	}

	// Install the signal and error handlers.
	err = afrouter.InitExitHandler()
	if err != nil {
		log.Errorf("Failed to initialize exit handler, exiting: %v", err)
		return 1
	}

	// Create the affinity router proxy...
	if ap, err := afrouter.NewArouterProxy(conf); err != nil {
		log.Errorf("Failed to create the arouter proxy, exiting:%v", err)
		return 1
		// and start it.
		// This function never returns unless an error
		// occurs or a signal is caught.
	} else if *conf.DryRun {
		// Do nothing
	} else if err := ap.ListenAndServe(); err != nil {
		log.Errorf("Exiting on error %v", err)
		return 1
	}

	return 0
}

func main() {
	status := startup()
	if status != 0 {
		os.Exit(status)
	}
}
