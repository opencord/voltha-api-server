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
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	"github.com/opencord/voltha-lib-go/v2/pkg/probe"
	pb "github.com/opencord/voltha-protos/v2/go/afrouter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
	"math"
	"os"
	"strconv"
	"time"
)

type volthaPod struct {
	name    string
	ipAddr  string
	node    string
	devIds  map[string]struct{}
	backend string
}

// TODO: These variables should be passed in from main() rather than
// declared here.

var (
	// if k8s variables are undefined, will attempt to use in-cluster config
	k8sApiServer      = GetStrEnv("K8S_API_SERVER", "")
	k8sKubeConfigPath = GetStrEnv("K8S_KUBE_CONFIG_PATH", "")

	podNamespace          = GetStrEnv("POD_NAMESPACE", "voltha")
	podLabelSelector      = GetStrEnv("POD_LABEL_SELECTOR", "app=rw-core")
	podAffinityGroupLabel = GetStrEnv("POD_AFFINITY_GROUP_LABEL", "affinity-group")

	podGrpcPort = uint64(GetIntEnv("POD_GRPC_PORT", 0, math.MaxUint16, 50057))

	afrouterRouterName    = GetStrEnv("AFROUTER_ROUTER_NAME", "vcore")
	afrouterRouteName     = GetStrEnv("AFROUTER_ROUTE_NAME", "dev_manager")
	afrouterRWClusterName = GetStrEnv("AFROUTER_RW_CLUSTER_NAME", "vcore")

	kafkaTopic      = GetStrEnv("KAFKA_TOPIC", "AffinityRouter")
	kafkaClientType = GetStrEnv("KAFKA_CLIENT_TYPE", "sarama")
	kafkaHost       = GetStrEnv("KAFKA_HOST", "kafka")
	kafkaPort       = GetIntEnv("KAFKA_PORT", 0, math.MaxUint16, 9092)
	kafkaInstanceID = GetStrEnv("KAFKA_INSTANCE_ID", "arouterd")
)

func GetIntEnv(key string, min, max, defaultValue int) int {
	if val, have := os.LookupEnv(key); have {
		num, err := strconv.Atoi(val)
		if err != nil || !(min <= num && num <= max) {
			panic(fmt.Errorf("%s must be a number in the range [%d, %d]; default: %d", key, min, max, defaultValue))
		}
		return num
	}
	return defaultValue
}

func GetStrEnv(key, defaultValue string) string {
	if val, have := os.LookupEnv(key); have {
		return val
	}
	return defaultValue
}

func Connect(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	log.Debugf("Trying to connect to %s", addr)
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithBackoffMaxDelay(time.Second*5),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: time.Second * 10, Timeout: time.Second * 5}))
	if err == nil {
		log.Debugf("Connection succeeded")
	}
	return conn, err
}

func setAffinity(ctx context.Context, client pb.ConfigurationClient, deviceId string, backend string) {
	log.Debugf("Configuring backend %s with device id %s \n", backend, deviceId)
	if res, err := client.SetAffinity(ctx, &pb.Affinity{
		Router:  afrouterRouterName,
		Route:   afrouterRouteName,
		Cluster: afrouterRWClusterName,
		Backend: backend,
		Id:      deviceId,
	}); err != nil {
		log.Debugf("failed affinity RPC call: %s\n", err)
	} else {
		log.Debugf("Result: %v\n", res)
	}
}

// endOnClose cancels the context when the connection closes
func ConnectionActiveContext(conn *grpc.ClientConn, p *probe.Probe) context.Context {
	ctx, disconnected := context.WithCancel(context.Background())
	go func() {
		for state := conn.GetState(); state != connectivity.TransientFailure && state != connectivity.Shutdown; state = conn.GetState() {
			if !conn.WaitForStateChange(context.Background(), state) {
				break
			}
		}
		log.Infof("Connection to afrouter lost")
		p.UpdateStatus("affinity-router", probe.ServiceStatusStopped)
		disconnected()
	}()
	return ctx
}
