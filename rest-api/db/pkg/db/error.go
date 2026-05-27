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

package db

import (
	"errors"
)

var (
	// ErrDoesNotExist is raised a DB query fails to find the requested entity
	ErrDoesNotExist = errors.New("the requested entity does not exist")
	// ErrDBError is a generalized error to expose to the user when unexpected errors occur when communicating with DB
	ErrDBError = errors.New("error communicating with data store")
	// ErrInvalidValue is raised when a value to be stored in DB is invalid
	ErrInvalidValue = errors.New("provided value is invalid")
	// ErrInvalidParams is raised when a function is called with invalid set of parameters
	ErrInvalidParams = errors.New("provided params are invalid or conflicting")

	// ErrXactAdvisoryLockFailed indicates that the transaction advisory lock could not be taken
	ErrXactAdvisoryLockFailed = errors.New("unable to take transaction advisory lock")
	// ErrSessionAdvisoryLockFailed indicates that the session advisory lock could not be taken
	ErrSessionAdvisoryLockFailed = errors.New("unable to take session advisory lock")
	// ErrSessionAdvisoryLockUnlockFailed indicates that the session advisory lock could not be released.
	ErrSessionAdvisoryLockUnlockFailed = errors.New("unable to release session advisory lock or lock was not held by this session")

	// ErrInvalidPort indicates the DB_PORT environment variable is not a valid integer.
	ErrInvalidPort = errors.New("failed to parse DB_PORT")
	// ErrInvalidCredential indicates the credential is not valid.
	ErrInvalidCredential = errors.New("invalid credential")

	// ErrTransactionInitiation is returned by WithTx*/WithTxResult* when the
	// underlying BeginTx call fails. HandleTxError detects this sentinel via
	// errors.Is and renders a user-facing message about transaction initiation.
	ErrTransactionInitiation = errors.New("DB transaction initiation error")
	// ErrTransactionCommit is returned by WithTx*/WithTxResult* when the
	// underlying tx.Commit call fails. HandleTxError detects this sentinel via
	// errors.Is and renders a user-facing message about transaction commit.
	ErrTransactionCommit = errors.New("DB transaction commit error")
)
