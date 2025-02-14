/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package triple

import (
	"context"
	"reflect"
	"sync"
)

import (
	perrors "github.com/pkg/errors"
	"google.golang.org/grpc"
)

import (
	"github.com/dubbogo/triple/internal/codec"
	"github.com/dubbogo/triple/internal/tools"
	"github.com/dubbogo/triple/pkg/common"
	"github.com/dubbogo/triple/pkg/common/constant"
	"github.com/dubbogo/triple/pkg/config"
)

// TripleClient client endpoint that using triple protocol
type TripleClient struct {
	h2Controller *H2Controller
	StubInvoker  reflect.Value

	//once is used when destroy
	once sync.Once

	// triple config
	opt *config.Option

	// serializer is triple serializer to do codec
	serializer common.Dubbo3Serializer
}

// NewTripleClient creates triple client
// it returns tripleClient, which contains invoker and triple connection.
// @impl must have method: GetDubboStub(cc *dubbo3.TripleConn) interface{}, to be capable with grpc
// @opt is used to init http2 controller, if it's nil, use the default config
func NewTripleClient(impl interface{}, opt *config.Option) (*TripleClient, error) {
	opt = tools.AddDefaultOption(opt)

	tripleClient := &TripleClient{
		opt: opt,
	}
	// start triple client connection,
	if err := tripleClient.connect(opt); err != nil {
		return nil, err
	}

	// put dubbo3 network logic to tripleConn, creat pb stub invoker
	if opt.SerializerType == constant.PBSerializerName {
		tripleClient.StubInvoker = reflect.ValueOf(getInvoker(impl, newTripleConn(tripleClient)))
	}

	return tripleClient, nil
}

// Connect called when new TripleClient, which start a tcp conn with target addr
func (t *TripleClient) connect(opt *config.Option) error {
	t.opt.Logger.Debugf("want to connect to location = ", opt.Location)

	var err error
	t.h2Controller, err = NewH2Controller(false, nil, t.opt)
	if err != nil {
		t.opt.Logger.Errorf("dubbo client new http2 controller error = %v", err)
		return err
	}
	t.h2Controller.address = opt.Location

	return nil
}

// Invoke call remote using stub
func (t *TripleClient) Invoke(methodName string, in []reflect.Value) []reflect.Value {
	rsp := make([]reflect.Value, 0, 2)
	switch t.opt.SerializerType {
	case constant.PBSerializerName:
		method := t.StubInvoker.MethodByName(methodName)
		// call function in pb.go
		return method.Call(in)
	case constant.TripleHessianWrapperSerializerName:
		out := codec.HessianUnmarshalStruct{}
		ctx := in[0].Interface().(context.Context)
		interfaceKey := ctx.Value(constant.InterfaceKey).(string)
		err := t.Request(ctx, "/"+interfaceKey+"/"+methodName, in[1].Interface(), &out)
		rsp = append(rsp, reflect.ValueOf(out.Val))
		if err != nil {
			return append(rsp, reflect.ValueOf(err))
		}
		return append(rsp, reflect.Value{})
	default:
		t.opt.Logger.Errorf("Invalid triple client serializerType = %s", t.opt.SerializerType)
		rsp = append(rsp, reflect.Value{})
		return append(rsp, reflect.ValueOf(perrors.Errorf("Invalid triple client serializerType = %s", t.opt.SerializerType)))
	}
}

// Request call h2Controller to send unary rpc req to server
// @path is /interfaceKey/functionName e.g. /com.apache.dubbo.sample.basic.IGreeter/BigUnaryTest
// @arg is request body
func (t *TripleClient) Request(ctx context.Context, path string, arg, reply interface{}) error {
	return t.h2Controller.UnaryInvoke(ctx, path, arg, reply)
}

// StreamRequest call h2Controller to send streaming request to sever, to start link.
// @path is /interfaceKey/functionName e.g. /com.apache.dubbo.sample.basic.IGreeter/BigStreamTest
func (t *TripleClient) StreamRequest(ctx context.Context, path string) (grpc.ClientStream, error) {
	return t.h2Controller.StreamInvoke(ctx, path)
}

// Close destroy http controller and return
func (t *TripleClient) Close() {
	t.opt.Logger.Debug("Triple Client Is closing")
	t.h2Controller.Destroy()
}

// IsAvailable returns if triple client is available
func (t *TripleClient) IsAvailable() bool {
	return t.h2Controller.IsAvailable()
}
