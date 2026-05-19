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

package client

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"

	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
)

// Errors
var (
	ErrFlowClientInvalidAddress    = errors.New("FlowClient: invalid address")
	ErrFlowClientInvalidDialOpts   = errors.New("FlowClient: invalid dial options")
	ErrFlowClientInvalidSecureOpts = errors.New("FlowClient: invalid secure options")
	ErrFlowClientInvalidServerCA   = errors.New("FlowClient: invalid server CA")
	ErrFlowClientInvalidClientCA   = errors.New("FlowClient: invalid client CA")
	ErrFlowClientInvalidClientKey  = errors.New("FlowClient: invalid client key")
	ErrFlowClientInvalidClientCert = errors.New("FlowClient: invalid client cert")
)

// SecureOptions is the enum for the secure options
type FlowClientSecureOptions int

const (
	// FlowInsecureGrpc is the insecure dial option
	FlowInsecureGrpc FlowClientSecureOptions = iota
	// FlowServerTLS is the secure dial option for server tls
	FlowServerTLS
	// FlowMutualTLS for mutual tls
	FlowMutualTLS

	// defaultCheckCertificateIntervalSeconds is the default interval to check for certificate changes
	defaultCheckFlowCertificateIntervalSeconds = 15 * 60 // 15 minutes in seconds
)

// FlowClientConfig is the data structure for the client configuration
type FlowClientConfig struct {
	// The address of the server <host>:<port>
	Address string
	// Secure flag
	Secure FlowClientSecureOptions
	// Skip Server Auth
	SkipServerAuth bool
	// The TLS certificate for the server
	ServerCAPath string
	// The TLS certificate for the client
	ClientCertPath string
	// The TLS key for the client
	ClientKeyPath string
	// client metrics interface
	ClientMetrics Metrics
}

// NewFlowClient creates a new FlowClient
func NewFlowClient(config *FlowClientConfig) (client *FlowClient, err error) {
	// Validate the config
	if config.Address == "" {
		log.Error().Err(ErrFlowClientInvalidAddress).Msg("FlowClient: no address provided")
		return nil, ErrFlowClientInvalidAddress
	}
	client = &FlowClient{}

	switch config.Secure {
	case FlowInsecureGrpc:
		// No secure options
		// Default option
		// connect with plain TCP
		log.Debug().Msg("FlowClient: insecure gRPC")
		client.dialOpts = append(client.dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	case FlowServerTLS:
		log.Debug().Msg("FlowClient: server TLS")
		// Validate the config contains server ca path
		if config.ServerCAPath == "" {
			log.Error().Err(ErrFlowClientInvalidServerCA).Msg("FlowClient: no server ca path provided")
			return nil, ErrFlowClientInvalidServerCA
		}
		if config.SkipServerAuth {
			// Server TLS
			// connect with TLS but not mutual TLS
			log.Info().Msg("FlowClient: skipping server auth in TLS ( Warn: This shouldn't be used in Prod)")
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
			}
			// Load the server ca
			_, err := credentials.NewClientTLSFromFile(config.ServerCAPath, "")
			if err != nil {
				log.Error().Err(err).Msg("FlowClient: failed to load server ca")
				return nil, err
			}

			// Create client dial option
			// Append the dial option
			client.dialOpts = append(client.dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

		} else {
			// Server TLS
			// connect with TLS but not mutual TLS
			// Load the server ca
			creds, err := credentials.NewClientTLSFromFile(config.ServerCAPath, "")
			if err != nil {
				log.Error().Err(err).Msg("FlowClient: failed to load server ca")
				return nil, err
			}
			// Append the dial option
			client.dialOpts = append(client.dialOpts, grpc.WithTransportCredentials(creds))
		}
	case FlowMutualTLS:
		// Mutual TLS
		// connect with mutual TLS
		log.Debug().Msg("FlowClient: mutual TLS")
		// 1. Load the client certificates
		clientCert, err := tls.LoadX509KeyPair(config.ClientCertPath, config.ClientKeyPath)
		if err != nil {
			log.Error().Err(err).Msg("FlowClient: failed to load client certificates")
			return nil, err
		}
		// 2. Load the Trust chain, root ca
		cabytes, err := os.ReadFile(config.ServerCAPath)
		if err != nil {
			log.Error().Err(err).Msg("FlowClient: failed to load Root CA certificates")

			return nil, err
		}
		capool := x509.NewCertPool()
		if !capool.AppendCertsFromPEM(cabytes) {
			return nil, fmt.Errorf("FlowClient: failed to append ca certificates to ca pool")
		}
		mutualTLSConfig := &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      capool,
		}
		creds := credentials.NewTLS(mutualTLSConfig)

		// Append to the dial option
		client.dialOpts = append(client.dialOpts, grpc.WithTransportCredentials(creds))

	default:
		log.Error().Err(ErrFlowClientInvalidSecureOpts).Msg("FlowClient: invalid dial options")
		return nil, ErrFlowClientInvalidSecureOpts
	}

	// configure interceptors
	var unaryInterceptors []grpc.UnaryClientInterceptor
	if config.ClientMetrics != nil {
		unaryInterceptors = append(unaryInterceptors, newGrpcUnaryMetricsInterceptor(config.ClientMetrics))
	}
	var streamInterceptors []grpc.StreamClientInterceptor
	if config.ClientMetrics != nil {
		streamInterceptors = append(streamInterceptors, newGrpcStreamMetricsInterceptor(config.ClientMetrics))
	}
	if os.Getenv("LS_SERVICE_NAME") != "" {
		handler := otelgrpc.NewClientHandler(otelgrpc.WithPropagators(otel.GetTextMapPropagator()))
		client.dialOpts = append(client.dialOpts, grpc.WithStatsHandler(handler))
	}
	if len(unaryInterceptors) > 0 {
		client.dialOpts = append(client.dialOpts, grpc.WithUnaryInterceptor(grpcmw.ChainUnaryClient(unaryInterceptors...)))
	}
	if len(streamInterceptors) > 0 {
		client.dialOpts = append(client.dialOpts, grpc.WithStreamInterceptor(grpcmw.ChainStreamClient(streamInterceptors...)))
	}

	// Create the client connection
	client.conn, err = grpc.NewClient(config.Address, client.dialOpts...)
	if err != nil {
		log.Error().Err(err).Msg("FlowClient: failed to initialize gRPC client")
		return nil, err
	}
	log.Info().Msg("FlowClient: gRPC client initialized")

	// Create Flow client
	client.flow = flowv1.NewFlowClient(client.conn)
	log.Info().Msg("FlowClient: client created")

	// Check the version of the server
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(5000)*time.Millisecond))
	defer cancel()
	_, err = client.flow.Version(ctx, &flowv1.VersionRequest{})
	if err != nil {
		log.Error().Err(err).Msg("FlowClient: failed to get version from server")
		return nil, fmt.Errorf("FlowClient: failed to get version from server: %w", err)
	}

	log.Info().Msg("FlowClient: successfully connected to server")

	return client, nil
}

// FlowClient is the data structure for the client
type FlowClient struct {
	// The client connection
	conn *grpc.ClientConn
	// gRPC dial options
	dialOpts []grpc.DialOption
	// flow client interface
	flow flowv1.FlowClient
}

// Close gracefully shuts down the client's gRPC connection.
func (cc *FlowClient) Close() error {
	if cc.conn != nil {
		// Close the grpc.ClientConn.
		return cc.conn.Close()
	}
	return nil
}

// Flow client getter
func (client *FlowClient) Flow() flowv1.FlowClient {
	return client.flow
}

// FlowAtomicClient is an atomic wrapper around the FlowClient
type FlowAtomicClient struct {
	Config  *FlowClientConfig
	value   *atomic.Value
	version atomic.Int64
}

// Version returns the current version of the FlowClient
func (rac *FlowAtomicClient) Version() int64 {
	return rac.version.Load()
}

// SwapClient atomically replaces the current FlowClient with a new one,
// returning the old client for the caller to manage.
func (rac *FlowAtomicClient) SwapClient(newClient *FlowClient) *FlowClient {

	// Atomically replace the current client with the new one and return the old client.
	oldClientInterface := rac.value.Swap(newClient)

	// Type assert the returned value to *FlowClient.
	// This should always succeed if the correct type was stored initially.
	oldClient, ok := oldClientInterface.(*FlowClient)
	if !ok {
		log.Error().Msg("SwapClient: Type assertion failed for the old client")
		return nil
	}

	// Increment the version number
	rac.version.Add(1)

	return oldClient
}

// GetClient returns the current version of Flow client from the atomic value.
// Returns nil if the client has not been initialized yet.
func (rac *FlowAtomicClient) GetClient() *FlowClient {
	v := rac.value.Load()
	if v == nil {
		return nil
	}
	client, _ := v.(*FlowClient)

	return client
}

// GetFlowClient returns the underlying Flow gRPC client. Returns ErrClientNotConnected
// if the client has not been initialized or is not currently connected.
// Prefer this over GetClient() + manual nil-check + .Flow() at call sites.
func (rac *FlowAtomicClient) GetFlowClient() (flowv1.FlowClient, error) {
	client := rac.GetClient()
	if client == nil {
		return nil, ErrClientNotConnected
	}
	// It's true that NewFlowClient always populates the inner flow field, BUT,
	// guard against zero-value FlowClient instances slipping in via direct
	// construction. Without this, a misconstructed wrapper would yield (nil,
	// nil) and break things.
	flow := client.Flow()
	if flow == nil {
		return nil, ErrClientNotConnected
	}
	return flow, nil
}

// CheckAndReloadCerts continuously monitors the TLS certificates for changes.
// If a change is detected, it reinitializes the FlowClient with the new certificates to ensure secure communication.
func (rac *FlowAtomicClient) CheckAndReloadCerts(initialClientCertMD5, initialServerCAMD5 []byte) {
	// Initialize contextual logger
	logger := log.With().Str("Component", "Flow").Str("Operation", "CheckAndReloadCerts").Logger()

	ticker := time.NewTicker(getFlowCertificateCheckInterval())
	defer ticker.Stop()

	lastClientCertMD5, lastServerCAMD5 := initialClientCertMD5, initialServerCAMD5

	for range ticker.C {
		changed, newClientMD5, newServerMD5, err := rac.CheckCertificates(lastClientCertMD5, lastServerCAMD5)
		if err != nil {
			logger.Error().Err(err).Msg("Error checking certificates for changes")
			continue
		}

		if changed {
			newClient, err := NewFlowClient(rac.Config)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to reinitialize gRPC client with new certificates")
				continue
			}

			// Atomically update the client instance and get the old one.
			oldClient := rac.SwapClient(newClient)

			// Delayed closure of the old client.
			go func(clientToClose *FlowClient) {
				// Delay the closure to allow ongoing client requests to complete.
				time.Sleep(10 * time.Second) // Adjust the delay as needed.

				// Ensure the client exists and has a connection to close.
				if clientToClose != nil {
					if err := clientToClose.Close(); err != nil {
						log.Error().Err(err).Msg("Error closing old FlowClient connection")
					}
				}
			}(oldClient)

			logger.Info().Msg("gRPC client successfully reinitialized with new certificates")

			// Update the stored MD5 hashes with the new ones for the next comparison.
			lastClientCertMD5, lastServerCAMD5 = newClientMD5, newServerMD5
		}
	}
}

// GetInitialCertMD5 retrieves the MD5 hash of the initial set of certificate that the client is Using
func (rac *FlowAtomicClient) GetInitialCertMD5() (clientCertMD5, serverCAMD5 []byte, err error) {
	// Load and hash the client certificate
	clientCertBytes, err := os.ReadFile(rac.Config.ClientCertPath)
	if err != nil {
		return nil, nil, err
	}
	clientCertMD5Hash := md5.Sum(clientCertBytes)
	clientCertMD5 = clientCertMD5Hash[:]

	// Load and hash the server CA certificate using os.ReadFile
	serverCABytes, err := os.ReadFile(rac.Config.ServerCAPath)
	if err != nil {
		return nil, nil, err
	}
	serverCAMD5Hash := md5.Sum(serverCABytes)
	serverCAMD5 = serverCAMD5Hash[:]

	return clientCertMD5, serverCAMD5, nil
}

// CheckCertificates checks if the client and server CA certificates have changed
func (rac *FlowAtomicClient) CheckCertificates(lastClientCertMD5, lastServerCAMD5 []byte) (bool, []byte, []byte, error) {
	// Load and hash the client certificate using os.ReadFile
	clientCertBytes, err := os.ReadFile(rac.Config.ClientCertPath)
	if err != nil {
		return false, lastClientCertMD5, lastServerCAMD5, err
	}
	clientCertMD5 := md5.Sum(clientCertBytes)

	// Load and hash the server CA certificate using os.ReadFile
	serverCABytes, err := os.ReadFile(rac.Config.ServerCAPath)
	if err != nil {
		return false, lastClientCertMD5, lastServerCAMD5, err
	}
	serverCAMD5 := md5.Sum(serverCABytes)

	// Check if either certificate has changed
	if !equalMD5(lastClientCertMD5, clientCertMD5[:]) || !equalMD5(lastServerCAMD5, serverCAMD5[:]) {
		return true, clientCertMD5[:], serverCAMD5[:], nil
	}

	return false, lastClientCertMD5, lastServerCAMD5, nil
}

// NewFlowAtomicClient creates a new FlowAtomicClient
func NewFlowAtomicClient(config *FlowClientConfig) *FlowAtomicClient {
	// Create the atomic value
	atomicClient := &FlowAtomicClient{
		Config:  config,
		value:   &atomic.Value{},
		version: atomic.Int64{},
	}

	return atomicClient
}

func getFlowCertificateCheckInterval() time.Duration {
	value, ok := os.LookupEnv("FLOW_CERT_CHECK_INTERVAL")
	if !ok {
		return defaultCheckFlowCertificateIntervalSeconds * time.Second
	}
	interval, err := strconv.Atoi(value)
	if err != nil {
		log.Error().Err(err).Str("FLOW_CERT_CHECK_INTERVAL", value).Msg("Invalid FLOW_CERT_CHECK_INTERVAL value; using default.")
		return defaultCheckFlowCertificateIntervalSeconds * time.Second
	}
	if interval <= 0 {
		log.Error().Int("FLOW_CERT_CHECK_INTERVAL", interval).Msg("FLOW_CERT_CHECK_INTERVAL must be > 0; using default.")
		return defaultCheckFlowCertificateIntervalSeconds * time.Second
	}
	return time.Duration(interval) * time.Second
}
