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
	"errors"
	"fmt"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
	"k8s.io/api/core/v1"
	"math"

	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha-go/common/log"
	"github.com/opencord/voltha-go/kafka"
	pb "github.com/opencord/voltha-protos/go/afrouter"
	cmn "github.com/opencord/voltha-protos/go/common"
	vpb "github.com/opencord/voltha-protos/go/voltha"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type volthaPod struct {
	name    string
	ipAddr  string
	node    string
	devIds  map[string]struct{}
	backend string
}

type Configuration struct {
	DisplayVersionOnly *bool
}

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

func newKafkaClient(clientType string, host string, port int, instanceID string) (kafka.Client, error) {
	log.Infow("kafka-client-type", log.Fields{"client": clientType})
	switch clientType {
	case "sarama":
		return kafka.NewSaramaClient(
			kafka.Host(host),
			kafka.Port(port),
			kafka.ConsumerType(kafka.GroupCustomer),
			kafka.ProducerReturnOnErrors(true),
			kafka.ProducerReturnOnSuccess(true),
			kafka.ProducerMaxRetries(6),
			kafka.NumPartitions(3),
			kafka.ConsumerGroupName(instanceID),
			kafka.ConsumerGroupPrefix(instanceID),
			kafka.AutoCreateTopic(false),
			kafka.ProducerFlushFrequency(5),
			kafka.ProducerRetryBackoff(time.Millisecond*30)), nil
	}
	return nil, errors.New("unsupported-client-type")
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
func ConnectionActiveContext(conn *grpc.ClientConn) context.Context {
	ctx, disconnected := context.WithCancel(context.Background())
	go func() {
		for state := conn.GetState(); state != connectivity.TransientFailure && state != connectivity.Shutdown; state = conn.GetState() {
			if !conn.WaitForStateChange(context.Background(), state) {
				break
			}
		}
		log.Infof("Connection to afrouter lost")
		disconnected()
	}()
	return ctx
}
