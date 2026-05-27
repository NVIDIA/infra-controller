/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package temporal

import (
	"errors"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"

	"github.com/NVIDIA/infra-controller-rest/common/pkg/endpoint"
)

const (
	defaultKeepAliveTime    = 10 * time.Second
	defaultKeepAliveTimeout = 60 * time.Second
)

type Client struct {
	config  Config
	options client.Options
	client  client.Client
}

type Config struct {
	Endpoint   endpoint.Config
	EnableTLS  bool
	ServerName string
	Namespace  string
}

func (c *Config) Validate() error {
	if err := c.Endpoint.Validate(); err != nil {
		return err
	}

	if c.EnableTLS {
		if c.ServerName == "" {
			return errors.New("server name is required")
		}

		if c.Endpoint.CACertificatePath == "" {
			return errors.New("CA certificate path is required")
		}

		if _, err := os.Stat(c.Endpoint.CACertificatePath); os.IsNotExist(err) { //nolint
			return errors.New("CA certificate path does not exist")
		}
	}

	return nil
}

func New(c Config) (*Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	tlsConfig, err := buildTLSConfig(c)
	if err != nil {
		return nil, err
	}

	options := client.Options{
		HostPort:  c.Endpoint.Target(),
		Namespace: c.Namespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS:              tlsConfig,
			KeepAliveTime:    defaultKeepAliveTime,
			KeepAliveTimeout: defaultKeepAliveTimeout,
		},
		DataConverter: converter.NewCompositeDataConverter(
			converter.NewNilPayloadConverter(),
			converter.NewByteSlicePayloadConverter(),
			converter.NewProtoJSONPayloadConverterWithOptions(
				converter.ProtoJSONPayloadConverterOptions{
					AllowUnknownFields: true,
				},
			),
			converter.NewProtoPayloadConverter(),
			converter.NewJSONPayloadConverter(),
		),
	}

	client, err := client.Dial(options)
	if err != nil {
		return nil, err
	}

	return &Client{config: c, options: options, client: client}, nil
}

func (c *Client) Client() client.Client {
	return c.client
}
