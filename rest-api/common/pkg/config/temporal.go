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

package config

import (
	"crypto/tls"
	"fmt"

	ctls "github.com/NVIDIA/infra-controller-rest/common/pkg/tls"
	cwfns "github.com/NVIDIA/infra-controller-rest/workflow/pkg/namespace"
)

// TemporalConfig holds configuration for Temporal communication
type TemporalConfig struct {
	Host          string
	Port          int
	ServerName    string
	Namespace     string
	Queue         string
	EncryptionKey string
	TLSEnabled    bool
	ClientTLSCfg  *tls.Config
	dynTLS        *ctls.DynTLSCfg
}

// GetHostPort returns the concatenated host & port
func (tcfg *TemporalConfig) GetHostPort() string {
	return fmt.Sprintf("%v:%v", tcfg.Host, tcfg.Port)
}

// Close cleans up TLS resources
func (tcfg *TemporalConfig) Close() {
	if tcfg.dynTLS != nil {
		tcfg.dynTLS.Close()
	}
}

// NewTemporalConfig initializes and returns a configuration object for managing Temporal
func NewTemporalConfig(host string, port int, serverName string, namespace string, queue string, encryptionKey string, tlsEnabled bool, certPath string, keyPath string, caPath string) (*TemporalConfig, error) {
	var dynTLS *ctls.DynTLSCfg
	var clientTLSCfg *tls.Config

	if tlsEnabled {
		var err error

		dynTLS, err = ctls.NewDynTLSCfg(keyPath, certPath, caPath)
		if err != nil {
			return nil, err
		}

		baseTLSCfg := &tls.Config{ //nolint:exhaustruct // only setting relevant TLS fields
			ServerName:         fmt.Sprintf("%s.%s", cwfns.CloudNamespace, serverName),
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}

		clientTLSCfg = dynTLS.WithTLSCfg(baseTLSCfg).ClientCfg()
	}

	return &TemporalConfig{
		Host:          host,
		Port:          port,
		Namespace:     namespace,
		Queue:         queue,
		ServerName:    serverName,
		EncryptionKey: encryptionKey,
		TLSEnabled:    tlsEnabled,
		ClientTLSCfg:  clientTLSCfg,
		dynTLS:        dynTLS,
	}, nil
}
