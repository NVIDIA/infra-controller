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

package componentmanager

import (
	"context"

	cmcatalog "github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/catalog"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/componentmanager/providerapi"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/executor/temporalworkflow/common"
	"github.com/NVIDIA/infra-controller-rest/flow/internal/task/operations"
	"github.com/NVIDIA/infra-controller-rest/flow/pkg/common/devicetypes"
)

type testManager struct {
	descriptor cmcatalog.Descriptor
}

func (m testManager) Descriptor() cmcatalog.Descriptor {
	return m.descriptor
}

func (m testManager) InjectExpectation(
	context.Context,
	common.Target,
	operations.InjectExpectationTaskInfo,
) error {
	return nil
}

func (m testManager) PowerControl(
	context.Context,
	common.Target,
	operations.PowerControlTaskInfo,
) error {
	return nil
}

func (m testManager) GetPowerStatus(
	context.Context,
	common.Target,
) (map[string]operations.PowerStatus, error) {
	return nil, nil
}

func (m testManager) FirmwareControl(
	context.Context,
	common.Target,
	operations.FirmwareControlTaskInfo,
) error {
	return nil
}

func (m testManager) GetFirmwareStatus(
	context.Context,
	common.Target,
) (map[string]operations.FirmwareUpdateStatus, error) {
	return nil, nil
}

func managerFactory(
	componentType devicetypes.ComponentType,
	implementation string,
	requiredProviders ...string,
) ManagerFactory {
	return func(*providerapi.ProviderRegistry) (ComponentManager, error) {
		return testManager{
			descriptor: testDescriptor(
				componentType,
				implementation,
				requiredProviders...,
			),
		}, nil
	}
}

func testDescriptor(
	componentType devicetypes.ComponentType,
	implementation string,
	requiredProviders ...string,
) cmcatalog.Descriptor {
	return cmcatalog.Descriptor{
		Type:              componentType,
		Implementation:    implementation,
		RequiredProviders: requiredProviders,
	}
}

func testFactorySpec(
	componentType devicetypes.ComponentType,
	implementation string,
	factory ManagerFactory,
	requiredProviders ...string,
) FactorySpec {
	return FactorySpec{
		Descriptor: testDescriptor(
			componentType,
			implementation,
			requiredProviders...,
		),
		Factory: factory,
	}
}
