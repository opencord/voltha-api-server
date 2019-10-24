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
	"github.com/golang/protobuf/ptypes"
	"github.com/opencord/voltha-lib-go/v2/pkg/kafka"
	"github.com/opencord/voltha-lib-go/v2/pkg/log"
	pb "github.com/opencord/voltha-protos/go/afrouter"
	ic "github.com/opencord/voltha-protos/go/inter_container"
	"golang.org/x/net/context"
	"regexp"
	"time"
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

func monitorDiscovery(kc kafka.Client, ctx context.Context, client pb.ConfigurationClient, ch <-chan *ic.InterContainerMessage, doneCh chan<- struct{}) {
	defer close(doneCh)
	defer kc.Stop()

monitorLoop:
	for {
		select {
		case <-ctx.Done():
			break monitorLoop
		case msg := <-ch:
			log.Debug("Received a device discovery notification")
			device := &ic.DeviceDiscovered{}
			if err := ptypes.UnmarshalAny(msg.Body, device); err != nil {
				log.Errorf("Could not unmarshal received notification %v", msg)
			} else {
				// somewhat hackish solution, backend is known from the first digit found in the publisher name
				group := regexp.MustCompile(`\d`).FindString(device.Publisher)
				if group != "" {
					// set the affinity of the discovered device
					setAffinity(ctx, client, device.Id, afrouterRWClusterName+group)
				} else {
					log.Error("backend is unknown")
				}
			}
		}
	}
}

func StartDiscoveryMonitor(ctx context.Context, client pb.ConfigurationClient) (<-chan struct{}, error) {
	doneCh := make(chan struct{})
	// Connect to kafka for discovery events
	kc, err := newKafkaClient(kafkaClientType, kafkaHost, kafkaPort, kafkaInstanceID)
	if err != nil {
		panic(err)
	}

	for {
		if err := kc.Start(); err != nil {
			log.Error("Could not connect to kafka")
		} else {
			break
		}
		select {
		case <-ctx.Done():
			close(doneCh)
			return doneCh, errors.New("GRPC context done")

		case <-time.After(5 * time.Second):
		}
	}
	ch, err := kc.Subscribe(&kafka.Topic{Name: kafkaTopic})
	if err != nil {
		log.Errorf("Could not subscribe to the '%s' channel, discovery disabled", kafkaTopic)
		close(doneCh)
		kc.Stop()
		return doneCh, err
	}

	go monitorDiscovery(kc, ctx, client, ch, doneCh)
	return doneCh, nil
}
