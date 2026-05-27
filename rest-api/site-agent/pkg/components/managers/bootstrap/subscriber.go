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

package bootstrap

import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// RegisterSubscriber registers Bootstrap workflows and activities with Temporal
func (api *BoostrapAPI) RegisterSubscriber() error {
	// Initialize logger
	logger := ManagerAccess.Data.EB.Log

	// Only master pod should watch for the OTP rotation workflow
	if !ManagerAccess.Conf.EB.IsMasterPod {
		return nil
	}

	// Register workflows
	wflowRegisterOptions := workflow.RegisterOptions{
		Name: "RotateTemporalCertAccessOTP",
	}
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterWorkflowWithOptions(api.RotateTemporalCertAccessOTP, wflowRegisterOptions)
	logger.Info().Msg("Bootstrap: Successfully registered RotateTemporalCertAccessOTP workflow")

	// Register activities
	otpHandler := NewOTPHandler(ManagerAccess.Data.EB.Managers.Bootstrap.Secret)

	activityRegisterOptions := activity.RegisterOptions{
		Name: "ReceiveAndSaveOTP",
	}
	ManagerAccess.Data.EB.Managers.Workflow.Temporal.Worker.RegisterActivityWithOptions(otpHandler.ReceiveAndSaveOTP, activityRegisterOptions)
	logger.Info().Msg("Bootstrap: Successfully registered ReceiveAndSaveOTP activity")

	return nil
}
