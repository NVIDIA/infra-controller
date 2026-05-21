#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Start a PostgreSQL container for local development and testing.
# This script creates a container that matches the CI environment.
#
# Usage:
#   ./scripts/start-test-postgres.sh        # Start the container
#   ./scripts/start-test-postgres.sh stop   # Stop and remove the container
#
# The container will be accessible at localhost:30432
# Credentials: postgres/postgres
# Database: flow_test

CONTAINER_NAME="flow-test-postgres"
POSTGRES_PORT=30432
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="postgres"
POSTGRES_DB="flow_test"

case "$1" in
    stop)
        echo "Stopping and removing container: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" 2>/dev/null
        docker rm "$CONTAINER_NAME" 2>/dev/null
        echo "Done."
        ;;
    status)
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            echo "Container $CONTAINER_NAME is running"
            docker ps --filter "name=$CONTAINER_NAME" --format "table {{.ID}}\t{{.Status}}\t{{.Ports}}"
        else
            echo "Container $CONTAINER_NAME is not running"
        fi
        ;;
    *)
        # Check if container is already running
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            echo "Container $CONTAINER_NAME is already running"
            docker ps --filter "name=$CONTAINER_NAME" --format "table {{.ID}}\t{{.Status}}\t{{.Ports}}"
            exit 0
        fi

        # Remove stopped container if exists
        docker rm "$CONTAINER_NAME" 2>/dev/null

        echo "Starting PostgreSQL container: $CONTAINER_NAME"
        # Durability is intentionally disabled: this container is for tests only
        # and the database is recreated for every test run.
        docker run -d \
            --name "$CONTAINER_NAME" \
            -p "${POSTGRES_PORT}:5432" \
            -e POSTGRES_USER="$POSTGRES_USER" \
            -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
            -e POSTGRES_DB="$POSTGRES_DB" \
            postgres:14.4-alpine \
            -c fsync=off \
            -c synchronous_commit=off \
            -c full_page_writes=off \
            -c wal_level=minimal \
            -c max_wal_senders=0

        echo ""
        echo "PostgreSQL is starting..."
        echo ""
        echo "Connection details:"
        echo "  Host:     localhost"
        echo "  Port:     $POSTGRES_PORT"
        echo "  User:     $POSTGRES_USER"
        echo "  Password: $POSTGRES_PASSWORD"
        echo "  Database: $POSTGRES_DB"
        echo ""
        echo "Environment variables for Flow:"
        echo "  export DB_ADDR=localhost"
        echo "  export DB_PORT=$POSTGRES_PORT"
        echo "  export DB_USER=$POSTGRES_USER"
        echo "  export DB_PASSWORD=$POSTGRES_PASSWORD"
        echo "  export DB_DATABASE=$POSTGRES_DB"
        echo ""
        echo "Run migrations:"
        echo "  go run . db migrate"
        echo ""
        echo "Run tests:"
        echo "  go test ./..."
        echo ""
        echo "To stop the container:"
        echo "  ./scripts/start-test-postgres.sh stop"
        ;;
esac
