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

package afrouter

import (
	"github.com/golang/protobuf/proto"
	"github.com/opencord/voltha-go/common/log"
	"google.golang.org/grpc/encoding"
)

/*
 * encoding.Codec renamed grpc.Codec's String() method to Name()
 *
 * grpc.ForceCodec() expects an encoding.Codec, alternative use of grpc.Codec is deprecated
 * grpc.CustomCodec() expects a grpc.Codec, has no mechanism to use encoding.Codec
 *
 * Make an interface that supports them both.
 */
type hybridCodec interface {
	encoding.Codec
	String() string
}

func Codec() hybridCodec {
	return CodecWithParent(&protoCodec{})
}

func CodecWithParent(parent encoding.Codec) hybridCodec {
	return &transparentRoutingCodec{parent}
}

type transparentRoutingCodec struct {
	parentCodec encoding.Codec
}

// responseFrame is a frame being "returned" to whomever established the connection
type responseFrame struct {
	payload []byte
	router  Router
	method  string
	backend *backend
	metaKey string
	metaVal string
}

// requestFrame is a frame coming in from whomever established the connection
type requestFrame struct {
	payload    []byte
	router     Router
	backend    *backend
	connection *connection // optional, if the router preferred one connection over another
	err        error
	methodInfo methodDetails
	serialNo   string
	metaKey    string
	metaVal    string
}

func (cdc *transparentRoutingCodec) Marshal(v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case *responseFrame:
		return t.payload, nil
	case *requestFrame:
		return t.payload, nil
	default:
		return cdc.parentCodec.Marshal(v)
	}
}

func (cdc *transparentRoutingCodec) Unmarshal(data []byte, v interface{}) error {
	switch t := v.(type) {
	case *responseFrame:
		t.payload = data
		// This is where the affinity is established on a northbound response

		/*
		 * NOTE: Ignoring this error is intentional. MethodRouter returns
		 *       error when a reply is not processed for northbound affinity,
		 *       but we still need to unmarshal it.
		 *
		 * TODO: Investigate error-return semantics of ReplyHandler.
		 */
		_ = t.router.ReplyHandler(v)
		return nil
	case *requestFrame:
		t.payload = data
		// This is were the afinity value is pulled from the payload
		// and the backend selected.
		t.backend, t.connection = t.router.Route(v)
		name := "<nil>"
		if t.backend != nil {
			name = t.backend.name
		}
		log.Debugf("Routing returned %s for method %s", name, t.methodInfo.method)

		return nil
	default:
		return cdc.parentCodec.Unmarshal(data, v)
	}
}

func (cdc *transparentRoutingCodec) Name() string {
	return cdc.parentCodec.Name()
}

func (cdc *transparentRoutingCodec) String() string {
	return cdc.Name()
}

// protoCodec is a Codec implementation with protobuf. It is the default Codec for gRPC.
type protoCodec struct{}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (protoCodec) Name() string {
	return "protoCodec"
}
