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

// Package types provides public domain types for the Flow client.
// This package has minimal dependencies (only uuid) and can be imported
// by external modules for interface definitions and mocking without
// pulling in gRPC dependencies.
package types

// ComponentType represents the type of a rack component.
type ComponentType string

const (
	ComponentTypeUnknown    ComponentType = "UNKNOWN"
	ComponentTypeCompute    ComponentType = "COMPUTE"
	ComponentTypeNVSwitch   ComponentType = "NVSWITCH"
	ComponentTypePowerShelf ComponentType = "POWERSHELF"
	ComponentTypeTORSwitch  ComponentType = "TORSWITCH"
	ComponentTypeUMS        ComponentType = "UMS"
	ComponentTypeCDU        ComponentType = "CDU"
)

// BMCType represents the type of BMC (Baseboard Management Controller).
type BMCType string

const (
	BMCTypeUnknown BMCType = "UNKNOWN"
	BMCTypeHost    BMCType = "HOST"
	BMCTypeDPU     BMCType = "DPU"
)

// PowerControlOp represents a power control operation.
type PowerControlOp string

const (
	PowerControlOpOn           PowerControlOp = "ON"
	PowerControlOpForceOn      PowerControlOp = "FORCE_ON"
	PowerControlOpOff          PowerControlOp = "OFF"
	PowerControlOpForceOff     PowerControlOp = "FORCE_OFF"
	PowerControlOpRestart      PowerControlOp = "RESTART"
	PowerControlOpForceRestart PowerControlOp = "FORCE_RESTART"
	PowerControlOpWarmReset    PowerControlOp = "WARM_RESET"
	PowerControlOpColdReset    PowerControlOp = "COLD_RESET"
)

// TaskStatus represents the status of an async task.
type TaskStatus string

const (
	TaskStatusUnknown   TaskStatus = "UNKNOWN"
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
)

// TaskExecutorType represents the type of task executor.
type TaskExecutorType string

const (
	TaskExecutorTypeUnknown  TaskExecutorType = "UNKNOWN"
	TaskExecutorTypeTemporal TaskExecutorType = "TEMPORAL"
)

// DiffType represents the type of difference in component validation.
type DiffType string

const (
	DiffTypeUnknown    DiffType = "Unknown"
	DiffTypeMissing    DiffType = "Missing"
	DiffTypeUnexpected DiffType = "Unexpected"
	DiffTypeDrift      DiffType = "Drift"
)

// OperationType represents the type of operation (power control, firmware, etc.).
type OperationType string

const (
	OperationTypeUnknown         OperationType = "UNKNOWN"
	OperationTypePowerControl    OperationType = "POWER_CONTROL"
	OperationTypeFirmwareControl OperationType = "FIRMWARE_CONTROL"
)
