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

package error

import (
	"go.temporal.io/sdk/temporal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrTypeInvalidRequest is returned when the request is invalid
	ErrTypeInvalidRequest         = "InvalidRequest"
	ErrTypeNICoObjectNotFound     = "NICoObjectNotFound"
	ErrTypeNICoUnimplemented      = "NICoUnimplemented"
	ErrTypeNICoUnavailable        = "NICoUnavailable"
	ErrTypeNICoDenied             = "NICoDenied"
	ErrTypeNICoAlreadyExists      = "NICoAlreadyExists"
	ErrTypeNICoFailedPrecondition = "NICoFailedPrecondition"
	ErrTypeNICoInvalidArgument    = "NICoInvalidArgument"

	// Legacy Carbide error type names. Retained so a newer REST can still
	// recognise errors emitted by an older site-workflow version that has
	// not yet been upgraded. Remove once the rollout window has closed.
	ErrTypeCarbideObjectNotFound     = "CarbideObjectNotFound"
	ErrTypeCarbideUnimplemented      = "CarbideUnimplemented"
	ErrTypeCarbideUnavailable        = "CarbideUnavailable"
	ErrTypeCarbideDenied             = "CarbideDenied"
	ErrTypeCarbideAlreadyExists      = "CarbideAlreadyExists"
	ErrTypeCarbideFailedPrecondition = "CarbideFailedPrecondition"
	ErrTypeCarbideInvalidArgument    = "CarbideInvalidArgument"
)

// ObjectNotFoundErrTypes returns the error types treated as "object not found"
// from a Site Agent gRPC call. Both NICo and legacy Carbide names are listed
// so that REST can recognise errors from older site-workflow deployments.
// Remove the Carbide entry once the rollout window has closed.
func ObjectNotFoundErrTypes() []string {
	return []string{ErrTypeNICoObjectNotFound, ErrTypeCarbideObjectNotFound}
}

// UnimplementedOrDeniedErrTypes returns the error types treated as
// "unimplemented or restricted" from a Site Agent gRPC call. Both NICo and
// legacy Carbide names are listed so that REST can recognise errors from
// older site-workflow deployments. Remove the Carbide entries once the
// rollout window has closed.
func UnimplementedOrDeniedErrTypes() []string {
	return []string{
		ErrTypeNICoUnimplemented,
		ErrTypeNICoDenied,
		ErrTypeCarbideUnimplemented,
		ErrTypeCarbideDenied,
	}
}

// FailedPreconditionErrTypes returns the error types treated as
// "failed precondition" from a Site Agent gRPC call. Both NICo and legacy
// Carbide names are listed so that REST can recognise errors from older
// site-workflow deployments. Remove the Carbide entry once the rollout
// window has closed.
func FailedPreconditionErrTypes() []string {
	return []string{ErrTypeNICoFailedPrecondition, ErrTypeCarbideFailedPrecondition}
}

// WrapError accepts an error and checks if it
// can be converted to a gRPC status.
//
// If the error can be converted and the status code matches a
// set of specific codes, it will be "wrapped" in a
// Temporal NewNonRetryableApplicationError.
//
// Otherwise, it returns the original error.
func WrapErr(err error) error {
	status, hasGrpcStatus := status.FromError(err)
	if hasGrpcStatus {
		switch status.Code() {
		case codes.NotFound:
			// If this is a 404 back from NICo, we'll bubble that back up as a custom temporal error.
			return temporal.NewNonRetryableApplicationError(err.Error(), ErrTypeNICoObjectNotFound, err)
		case codes.Unimplemented:
			return temporal.NewNonRetryableApplicationError(err.Error(), ErrTypeNICoUnimplemented, err)
		case codes.Unavailable:
			return temporal.NewNonRetryableApplicationError(err.Error(), ErrTypeNICoUnavailable, err)
		case codes.PermissionDenied:
			return temporal.NewNonRetryableApplicationError(err.Error(), ErrTypeNICoDenied, err)
		case codes.AlreadyExists:
			return temporal.NewNonRetryableApplicationError(err.Error(), ErrTypeNICoAlreadyExists, err)
		case codes.FailedPrecondition:
			return temporal.NewNonRetryableApplicationError(err.Error(), ErrTypeNICoFailedPrecondition, err)
		case codes.InvalidArgument:
			return temporal.NewNonRetryableApplicationError(err.Error(), ErrTypeNICoInvalidArgument, err)
		}
	}
	return err
}
