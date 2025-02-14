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

package config

import (
	"github.com/dubbogo/triple/pkg/common/constant"
	"github.com/dubbogo/triple/pkg/common/logger"
	"github.com/dubbogo/triple/pkg/common/logger/default_logger"
)

// triple option
type Option struct {
	// network opts
	Timeout    uint32
	BufferSize uint32

	// service opts
	Location       string
	Protocol       string
	SerializerType constant.TripleSerializerName

	// triple header opts
	HeaderGroup      string
	HeaderAppVersion string

	// logger
	Logger logger.Logger
}

// Validate sets empty field to default config
func (o *Option) Validate() {
	if o.Timeout == uint32(0) {
		o.Timeout = uint32(constant.DefaultTimeout)
	}

	if o.BufferSize == uint32(0) {
		o.BufferSize = uint32(constant.DefaultHttp2ControllerReadBufferSize)
	}

	if o.Location == "" {
		o.Location = constant.DefaultListeningAddress
	}

	if o.Logger == nil {
		o.Logger = default_logger.GetDefaultLogger()
	}

	if o.Protocol == "" {
		o.Protocol = constant.TRIPLE
	}

	if o.SerializerType == "" {
		o.SerializerType = constant.PBSerializerName
	}
}

// nolint
type OptionFunction func(o *Option)

// NewTripleOption return Triple Option with given config defined by @fs
func NewTripleOption(fs ...OptionFunction) *Option {
	opt := &Option{}
	for _, v := range fs {
		v(opt)
	}

	return opt
}

// WithClientTimeout return OptionFunction with timeout of @timeout
func WithClientTimeout(timeout uint32) OptionFunction {
	return func(o *Option) {
		o.Timeout = timeout
	}
}

// WithBufferSize return OptionFunction with buffer read size of @size
func WithBufferSize(size uint32) OptionFunction {
	return func(o *Option) {
		o.BufferSize = size
	}
}

// WithSerializerType return OptionFunction with target @serializerType, now we support "protobuf" and "hessian2"
func WithSerializerType(serializerType constant.TripleSerializerName) OptionFunction {
	return func(o *Option) {
		o.SerializerType = serializerType
	}
}

// WithProtocol return OptionFunction with target @protocol, now we support "tri"
func WithProtocol(protocol string) OptionFunction {
	return func(o *Option) {
		o.Protocol = protocol
	}
}

// WithLocation return OptionFunction with target @location, for example "127.0.0.1:20001"
func WithLocation(location string) OptionFunction {
	return func(o *Option) {
		o.Location = location
	}
}

// WithHeaderAppVersion return OptionFunction with target @appVersion, for example "1.0.0"
func WithHeaderAppVersion(appVersion string) OptionFunction {
	return func(o *Option) {
		o.HeaderAppVersion = appVersion
	}
}

// WithHeaderGroup return OptionFunction with target @group, for example "dubbogo"
func WithHeaderGroup(group string) OptionFunction {
	return func(o *Option) {
		o.HeaderGroup = group
	}
}

// WithLogger return OptionFunction with target @logger, which must impl triple/pkg/common/logger.Logger
func WithLogger(logger logger.Logger) OptionFunction {
	return func(o *Option) {
		o.Logger = logger
	}
}
