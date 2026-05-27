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

package simple

import (
	"context"

	"github.com/rs/zerolog"
)

type loggerKey struct{}

// Logger is an alias for *zerolog.Logger
type Logger = *zerolog.Logger

// NewNoOpLogger returns a no-op logger that discards all log messages
func NewNoOpLogger() Logger {
	logger := zerolog.Nop()
	return &logger
}

// WithLogger returns a new context with the given logger embedded
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerFromContext extracts the logger from the context.
// If no logger is found in the context, it returns a no-op logger.
func LoggerFromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey{}).(Logger); ok && logger != nil {
		return logger
	}
	return NewNoOpLogger()
}
