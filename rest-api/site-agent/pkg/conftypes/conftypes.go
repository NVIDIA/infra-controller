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

package conftypes

import (
	"encoding/json"
	"time"

	"github.com/NVIDIA/infra-controller-rest/site-workflow/pkg/grpc/client"
)

// RunInEnvironment provides a strongly-typed indicator for the environment
// in which the app is running.
type RunInEnvironment int

const (
	// RunningInUnknown - Running In Unknown Env
	RunningInUnknown RunInEnvironment = iota
	// RunningInDocker - Running In Docker
	RunningInDocker
	// RunningInK8s - Running In K8s
	RunningInK8s
)

// TemporalConfig holds configurations for connecting to Temporal server
type TemporalConfig struct {
	Host                       string `json:"host"`
	Port                       string `json:"port"`
	ClusterID                  string `json:"clusterID"`
	TemporalServer             string `json:"temporalServer"`
	TemporalPublishNamespace   string `json:"temporalPublishNamespace"`
	TemporalSubscribeNamespace string `json:"temporalSubscribeNamespace"`
	TemporalPublishQueue       string `json:"temporalPublishQueue"`
	TemporalSubscribeQueue     string `json:"temporalSubscribeQueue"`
	TemporalInventorySchedule  string `json:"temporalInventorySchedule"`
	TemporalCertPath           string `json:"temporalCertPath"`
}

// GetTemporalCertOTPFullPath - Get Temporal Cert OTP path
func (tc *TemporalConfig) GetTemporalCertOTPFullPath() string {
	return tc.TemporalCertPath + "/otp"
}

// GetTemporalCACertFilePath - Get Temporal CA Cert File n Path
func (tc *TemporalConfig) GetTemporalCACertFilePath() (string, string) {
	file := "ca.crt"
	return file, tc.TemporalCertPath + "/ca/"
}

// GetTemporalCACertFullPath - Get Temporal CA Cert Full Path
func (tc *TemporalConfig) GetTemporalCACertFullPath() string {
	file := "tls.crt"
	return tc.TemporalCertPath + "/ca/" + file
}

// GetTemporalClientCertFilePath - Get Temporal client Cert File n Path
func (tc *TemporalConfig) GetTemporalClientCertFilePath() ([]string, string) {
	file := []string{"tls.crt", "tls.key"}
	return file, tc.TemporalCertPath + "/client/"
}

// GetTemporalClientCertFullPath - Get Temporal client Cert Full Path
func (tc *TemporalConfig) GetTemporalClientCertFullPath() string {
	file := "tls.crt"
	return tc.TemporalCertPath + "/client/" + file
}

// GetTemporalClientKeyFullPath - Get Temporal client Key Full Path
func (tc *TemporalConfig) GetTemporalClientKeyFullPath() string {
	file := "tls.key"
	return tc.TemporalCertPath + "/client/" + file
}

// NICoConfig holds configurations for connecting to NICo server
type NICoConfig struct {
	Address        string               `json:"nicoAddress"`
	Secure         client.SecureOptions `json:"nicoSecureOptions"`
	SkipServerAuth bool                 `json:"nicoSkipServerAuth"`
	ServerCAPath   string               `json:"nicoCertPath"`
	ClientCertPath string               `json:"nicoClientCertPath"`
	ClientKeyPath  string               `json:"nicoClientKeyPath"`
}

// FlowConfig holds configurations for connecting to Flow server
type FlowConfig struct {
	Enabled        bool                           `json:"flowEnabled"`
	Address        string                         `json:"flowAddress"`
	Secure         client.FlowClientSecureOptions `json:"flowSecureOptions"`
	SkipServerAuth bool                           `json:"flowSkipServerAuth"`
	ServerCAPath   string                         `json:"flowCertPath"`
	ClientCertPath string                         `json:"flowClientCertPath"`
	ClientKeyPath  string                         `json:"flowClientKeyPath"`
}

// Config for Site Agent
type Config struct {
	Temporal         TemporalConfig
	NICo             NICoConfig
	Flow             FlowConfig
	IsMasterPod      bool          `json:"isMasterPod"`
	EnableDebug      bool          `json:"enableDebug"`
	DevMode          bool          `json:"devMode"`
	EnableTLS        bool          `json:"enableTLS"`
	DisableBootstrap bool          `json:"disableBootstrap"`
	BootstrapSecret  string        `json:"bootstrapSecret"` // Path to the bootstrap secret file
	WatcherInterval  time.Duration `json:"watcherInterval"`
	PodNamespace     string        `json:"podNamespace"`
	TemporalSecret   string        `json:"temporalSecret"`
	MetricsPort      string        `json:"metricsPort"`
	SiteVersion      string        `json:"siteVersion"`
	CloudVersion     string        `json:"cloudVersion"`
	RunningIn        RunInEnvironment
	UtMode           bool
}

// String - json string
func (c *Config) String() string {
	str, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(str)
}

// NewConfType - new config
func NewConfType() *Config {
	// We can set the default config here
	return &Config{
		Temporal: TemporalConfig{},
		NICo:     NICoConfig{},
		Flow:     FlowConfig{},
	}
}
