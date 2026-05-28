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

//! Admin UI live log viewer.
//!
//! `page()` serves up the unified logs hub. `stream()` is a Server-Sent Events
//! endpoint that replays the recent in-process tracing buffer and then tails live
//! events from [`crate::logging::stream::LogStream`]. Only the `api` source is
//! wired in for now. The idea is this logs hub will eventually be a place where
//! we can get similar log streaming for Scout and DPU agents via ScoutStream.

use std::convert::Infallible;
use std::sync::Arc;

use askama::Template;
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::response::sse::{Event, KeepAlive, Sse};
use axum::response::{Html, IntoResponse, Response};
use futures::stream::{self, StreamExt};
use tokio::sync::broadcast::error::RecvError;

use super::Base;
use crate::api::Api;
use crate::logging::stream::LogLine;

#[derive(Template)]
#[template(path = "api_logs.html")]
struct LogsPage {}

impl Base for LogsPage {}

/// `GET /admin/logs` — the unified live log viewer hub.
pub async fn page() -> Html<String> {
    Html(LogsPage {}.render().unwrap())
}

/// Handle `GET /admin/logs/{source}/stream`, which opens up
/// the Server-Sent Events stream of nico-api log lines.
pub async fn stream(State(state): State<Arc<Api>>, Path(source): Path<String>) -> Response {
    if source != "api" {
        return (
            StatusCode::NOT_FOUND,
            format!("log source {source:?} is not available yet"),
        )
            .into_response();
    }

    let log_stream = state.dynamic_settings.log_stream.clone();
    // Subscribe before snapshotting the backlog so no line slips through the gap
    // between the two.
    let rx = log_stream.subscribe();
    let backlog = log_stream.recent();

    let replay = stream::iter(
        backlog
            .into_iter()
            .map(|line| Ok::<_, Infallible>(line_event(line.as_ref()))),
    );

    let live = stream::unfold(rx, |mut rx| async move {
        match rx.recv().await {
            Ok(line) => Some((Ok::<_, Infallible>(line_event(line.as_ref())), rx)),
            // A subscriber that fell behind the bounded channel: tell the viewer
            // how many lines it missed rather than dropping the connection.
            Err(RecvError::Lagged(skipped)) => {
                let ev = Event::default().event("lag").data(skipped.to_string());
                Some((Ok(ev), rx))
            }
            Err(RecvError::Closed) => None,
        }
    });

    Sse::new(replay.chain(live))
        .keep_alive(KeepAlive::default())
        .into_response()
}

/// Serialize a log line into an SSE data frame (one JSON object per event).
fn line_event(line: &LogLine) -> Event {
    Event::default().data(serde_json::to_string(line).unwrap_or_default())
}
