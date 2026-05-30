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

use std::borrow::Cow;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::Instant;

use prometheus::{CounterVec, Gauge, Opts};

use super::client::typed_value_to_string;
use super::proto::{self, PathElem};
use super::sample_processor::now_unix_secs;
use super::subscriber::GnmiStreamMetrics;
use crate::HealthError;
use crate::sink::{CollectorEvent, DataSink, EventContext, SensorHealthData};

type ParsedRow = HashMap<String, String>;
type TableSnapshot = HashMap<String, ParsedRow>;

pub(crate) const ON_CHANGE_STREAM_ID_SYSTEM_EVENTS: &str = "nvue_gnmi_events";

pub(crate) struct OnChangeStreamMetrics {
    pub(crate) rows_total: CounterVec,
    pub(crate) last_row_timestamp: Gauge,
}

impl OnChangeStreamMetrics {
    pub(crate) fn new(
        registry: &prometheus::Registry,
        prefix: &str,
        stream_id: &str,
        const_labels: HashMap<String, String>,
    ) -> Result<Self, HealthError> {
        let rows_total = CounterVec::new(
            Opts::new(
                format!("{prefix}_{stream_id}_total"),
                "ON_CHANGE rows received by severity (field 'severity' if present)",
            )
            .const_labels(const_labels.clone()),
            &["severity"],
        )?;
        registry.register(Box::new(rows_total.clone()))?;

        let last_row_timestamp = Gauge::with_opts(
            Opts::new(
                format!("{prefix}_{stream_id}_last_timestamp"),
                "Unix timestamp of most recent ON_CHANGE row",
            )
            .const_labels(const_labels),
        )?;
        registry.register(Box::new(last_row_timestamp.clone()))?;

        Ok(Self {
            rows_total,
            last_row_timestamp,
        })
    }
}

pub(crate) struct GnmiOnChangeProcessor {
    pub(crate) collector_name: String,
    pub(crate) stream_metrics: OnChangeStreamMetrics,
    pub(crate) data_sink: Option<Arc<dyn DataSink>>,
    pub(crate) event_context: EventContext,
    pub(crate) switch_id: String,
    previous_snapshot: Mutex<TableSnapshot>,
}

impl GnmiOnChangeProcessor {
    pub(crate) fn new(
        collector_name: String,
        stream_metrics: OnChangeStreamMetrics,
        data_sink: Option<Arc<dyn DataSink>>,
        event_context: EventContext,
        switch_id: String,
    ) -> Self {
        Self {
            collector_name,
            stream_metrics,
            data_sink,
            event_context,
            switch_id,
            previous_snapshot: Mutex::new(HashMap::new()),
        }
    }

    #[allow(deprecated)]
    pub(crate) fn process_subscribe_response(
        &self,
        resp: &proto::SubscribeResponse,
        stream_metrics: &GnmiStreamMetrics,
    ) {
        let notification = match &resp.response {
            Some(proto::subscribe_response::Response::Update(n)) => n,
            Some(proto::subscribe_response::Response::SyncResponse(_)) => return,
            Some(proto::subscribe_response::Response::Error(e)) => {
                stream_metrics.stream_errors_total.inc();
                tracing::warn!(
                    code = e.code,
                    message = %e.message,
                    stream = %self.collector_name,
                    "nvue_gnmi ON_CHANGE: server error in stream"
                );
                return;
            }
            None => return,
        };

        stream_metrics.notifications_received_total.inc();
        stream_metrics
            .last_notification_timestamp
            .set(now_unix_secs());

        let start = Instant::now();
        let entity_count = self.process_notification(notification);
        stream_metrics
            .notification_processing_seconds
            .observe(start.elapsed().as_secs_f64());
        stream_metrics.monitored_entities.set(entity_count as f64);
    }

    fn process_notification(&self, notification: &proto::Notification) -> usize {
        let prefix_elems: &[PathElem] = notification
            .prefix
            .as_ref()
            .map(|p| p.elem.as_slice())
            .unwrap_or_default();

        let mut current: TableSnapshot = HashMap::new();

        for update in &notification.update {
            let val = match update.val.as_ref() {
                Some(v) => v,
                None => continue,
            };

            let update_elems: &[PathElem] = update
                .path
                .as_ref()
                .map(|p| p.elem.as_slice())
                .unwrap_or_default();

            let combined: Vec<&PathElem> = prefix_elems.iter().chain(update_elems.iter()).collect();

            let Some(instance_key) = find_instance_key(&combined) else {
                continue;
            };
            let Some(leaf_elem) = combined.last() else {
                continue;
            };

            let value = typed_value_to_string(val).unwrap_or_default();
            current
                .entry(instance_key.to_string())
                .or_default()
                .insert(leaf_elem.name.clone(), value);
        }

        let mut previous = match self.previous_snapshot.lock() {
            Ok(guard) => guard,
            Err(poisoned) => poisoned.into_inner(),
        };
        for (instance_id, row) in &current {
            let is_new_or_changed = previous.get(instance_id).map(|p| p != row).unwrap_or(true);
            if is_new_or_changed {
                self.emit_row_as_metric(instance_id, row);
            }
        }

        let entity_count = current.len();
        *previous = current;
        entity_count
    }

    fn emit_row_as_metric(&self, instance_id: &str, row: &ParsedRow) {
        let severity = row.get("severity").map(String::as_str).unwrap_or("unknown");
        let text = row.get("text").map(String::as_str).unwrap_or("");

        self.stream_metrics.last_row_timestamp.set(now_unix_secs());
        self.stream_metrics
            .rows_total
            .with_label_values(&[severity])
            .inc();

        tracing::info!(
            switch_id = %self.switch_id,
            stream = %self.collector_name,
            instance_id,
            severity,
            text,
            "nvue_gnmi ON_CHANGE: row received"
        );

        let Some(sink) = &self.data_sink else { return };

        let key = format!("{}:{}", self.collector_name, instance_id);
        let mut labels = vec![
            (Cow::Borrowed("instance_id"), instance_id.to_string()),
            (Cow::Borrowed("text"), text.to_string()),
        ];
        for (key, value) in row {
            if key != "text" {
                labels.push((Cow::Owned(key.clone()), value.clone()));
            }
        }

        sink.handle_event(
            &self.event_context,
            &CollectorEvent::Metric(Box::new(SensorHealthData {
                key,
                name: self.collector_name.clone(),
                metric_type: "on_change_row".to_string(),
                unit: "severity".to_string(),
                value: severity_to_f64(Some(severity)),
                labels,
                context: None,
            })),
        );
    }
}

fn find_instance_key<'a>(elems: &[&'a PathElem]) -> Option<&'a str> {
    elems
        .iter()
        .find(|e| !e.key.is_empty())
        .and_then(|e| e.key.values().next().map(String::as_str))
}

fn severity_to_f64(severity: Option<&str>) -> f64 {
    match severity {
        Some(s) if s.eq_ignore_ascii_case("informational") => 1.0,
        Some(s) if s.eq_ignore_ascii_case("warning") => 2.0,
        Some(s) if s.eq_ignore_ascii_case("error") => 3.0,
        Some(s) if s.eq_ignore_ascii_case("critical") => 4.0,
        _ => 0.0,
    }
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use carbide_uuid::rack::RackId;
    use carbide_uuid::switch::{SwitchId, SwitchIdSource, SwitchType};
    use mac_address::MacAddress;

    use super::*;
    use crate::endpoint::{BmcAddr, EndpointMetadata, SwitchData, SwitchEndpointRole};

    const TEST_COLLECTOR_NAME: &str = "nvue_gnmi_system_events";

    #[derive(Default)]
    struct CapturingSink {
        events: Mutex<Vec<(EventContext, CollectorEvent)>>,
    }

    impl DataSink for CapturingSink {
        fn sink_type(&self) -> &'static str {
            "capturing_sink"
        }

        fn handle_event(&self, context: &EventContext, event: &CollectorEvent) {
            self.events
                .lock()
                .expect("lock poisoned")
                .push((context.clone(), event.clone()));
        }
    }

    fn test_labels() -> HashMap<String, String> {
        HashMap::from([(
            "collector_type".to_string(),
            ON_CHANGE_STREAM_ID_SYSTEM_EVENTS.to_string(),
        )])
    }

    fn test_switch_id(label: &str) -> SwitchId {
        let mut hash = [0u8; 32];
        let bytes = label.as_bytes();
        hash[..bytes.len().min(32)].copy_from_slice(&bytes[..bytes.len().min(32)]);
        SwitchId::new(SwitchIdSource::Tpm, hash, SwitchType::NvLink)
    }

    fn test_event_context(collector_type: &'static str) -> EventContext {
        EventContext {
            endpoint_key: "aa:bb:cc:dd:ee:ff".to_string(),
            addr: BmcAddr {
                ip: "10.0.0.1".parse().unwrap(),
                port: None,
                mac: MacAddress::from_str("AA:BB:CC:DD:EE:FF").unwrap(),
            },
            collector_type,
            metadata: None,
            rack_id: None,
        }
    }

    fn test_processor(data_sink: Option<Arc<dyn DataSink>>) -> GnmiOnChangeProcessor {
        let registry = prometheus::Registry::new();
        let stream_metrics =
            OnChangeStreamMetrics::new(&registry, "test", TEST_COLLECTOR_NAME, test_labels())
                .unwrap();
        GnmiOnChangeProcessor::new(
            TEST_COLLECTOR_NAME.to_string(),
            stream_metrics,
            data_sink,
            test_event_context(TEST_COLLECTOR_NAME),
            "SN1234".to_string(),
        )
    }

    fn make_path_elem(name: &str, keys: &[(&str, &str)]) -> PathElem {
        PathElem {
            name: name.to_string(),
            key: keys
                .iter()
                .map(|(key, value)| (key.to_string(), value.to_string()))
                .collect(),
        }
    }

    fn make_typed_value_string(value: &str) -> proto::TypedValue {
        proto::TypedValue {
            value: Some(proto::typed_value::Value::StringVal(value.to_string())),
        }
    }

    #[test]
    fn test_find_instance_key() {
        let elems = [
            make_path_elem("system-events", &[]),
            make_path_elem("system-event", &[("event-id", "38")]),
            make_path_elem("state", &[]),
            make_path_elem("severity", &[]),
        ];
        let refs: Vec<&PathElem> = elems.iter().collect();
        assert_eq!(find_instance_key(&refs), Some("38"));
    }

    #[test]
    fn test_find_instance_key_missing() {
        let elems = [
            make_path_elem("system-events", &[]),
            make_path_elem("state", &[]),
        ];
        let refs: Vec<&PathElem> = elems.iter().collect();
        assert_eq!(find_instance_key(&refs), None);
    }

    #[test]
    fn test_severity_to_f64() {
        assert_eq!(severity_to_f64(Some("informational")), 1.0);
        assert_eq!(severity_to_f64(Some("warning")), 2.0);
        assert_eq!(severity_to_f64(Some("error")), 3.0);
        assert_eq!(severity_to_f64(Some("critical")), 4.0);
        assert_eq!(severity_to_f64(Some("CRITICAL")), 4.0);
        assert_eq!(severity_to_f64(Some("other")), 0.0);
        assert_eq!(severity_to_f64(None), 0.0);
    }

    #[test]
    fn test_on_change_stream_metrics_duplicate_registration_fails() {
        let registry = prometheus::Registry::new();
        let _ = OnChangeStreamMetrics::new(&registry, "test", "stream_a", test_labels()).unwrap();
        let result = OnChangeStreamMetrics::new(&registry, "test", "stream_a", test_labels());
        assert!(result.is_err());
    }

    #[test]
    fn test_process_notification_severity_and_text() {
        let processor = test_processor(None);
        let notification = proto::Notification {
            prefix: Some(proto::Path {
                elem: vec![make_path_elem("system-events", &[])],
                ..Default::default()
            }),
            update: vec![
                proto::Update {
                    path: Some(proto::Path {
                        elem: vec![
                            make_path_elem("system-event", &[("event-id", "5")]),
                            make_path_elem("state", &[]),
                            make_path_elem("severity", &[]),
                        ],
                        ..Default::default()
                    }),
                    val: Some(make_typed_value_string("critical")),
                    ..Default::default()
                },
                proto::Update {
                    path: Some(proto::Path {
                        elem: vec![
                            make_path_elem("system-event", &[("event-id", "5")]),
                            make_path_elem("state", &[]),
                            make_path_elem("text", &[]),
                        ],
                        ..Default::default()
                    }),
                    val: Some(make_typed_value_string("System fatal state detected")),
                    ..Default::default()
                },
            ],
            ..Default::default()
        };

        let count = processor.process_notification(&notification);
        assert_eq!(count, 1);
        assert_eq!(
            processor
                .stream_metrics
                .rows_total
                .with_label_values(&["critical"])
                .get(),
            1.0
        );
        assert!(processor.stream_metrics.last_row_timestamp.get() > 0.0);
    }

    #[test]
    fn test_process_notification_snapshot_diff_no_duplicate_emit() {
        let processor = test_processor(None);
        let notification = proto::Notification {
            prefix: Some(proto::Path {
                elem: vec![make_path_elem("system-events", &[])],
                ..Default::default()
            }),
            update: vec![
                proto::Update {
                    path: Some(proto::Path {
                        elem: vec![
                            make_path_elem("system-event", &[("event-id", "7")]),
                            make_path_elem("state", &[]),
                            make_path_elem("severity", &[]),
                        ],
                        ..Default::default()
                    }),
                    val: Some(make_typed_value_string("error")),
                    ..Default::default()
                },
                proto::Update {
                    path: Some(proto::Path {
                        elem: vec![
                            make_path_elem("system-event", &[("event-id", "7")]),
                            make_path_elem("state", &[]),
                            make_path_elem("text", &[]),
                        ],
                        ..Default::default()
                    }),
                    val: Some(make_typed_value_string("same event")),
                    ..Default::default()
                },
            ],
            ..Default::default()
        };

        processor.process_notification(&notification);
        processor.process_notification(&notification);

        assert_eq!(
            processor
                .stream_metrics
                .rows_total
                .with_label_values(&["error"])
                .get(),
            1.0
        );
    }

    #[test]
    fn emitted_metrics_preserve_switch_position_context() {
        let sink = Arc::new(CapturingSink::default());
        let switch_id = test_switch_id("switch-a");
        let registry = prometheus::Registry::new();
        let stream_metrics =
            OnChangeStreamMetrics::new(&registry, "test", TEST_COLLECTOR_NAME, test_labels())
                .unwrap();
        let processor = GnmiOnChangeProcessor::new(
            TEST_COLLECTOR_NAME.to_string(),
            stream_metrics,
            Some(sink.clone()),
            EventContext {
                endpoint_key: "aa:bb:cc:dd:ee:ff".to_string(),
                addr: BmcAddr {
                    ip: "10.0.0.1".parse().unwrap(),
                    port: None,
                    mac: MacAddress::from_str("AA:BB:CC:DD:EE:FF").unwrap(),
                },
                collector_type: ON_CHANGE_STREAM_ID_SYSTEM_EVENTS,
                metadata: Some(EndpointMetadata::Switch(SwitchData {
                    id: Some(switch_id),
                    serial: "SN-SWITCH-001".to_string(),
                    slot_number: Some(7),
                    tray_index: Some(3),
                    endpoint_role: SwitchEndpointRole::Host,
                    is_primary: false,
                    nmxt_enabled: false,
                })),
                rack_id: Some(RackId::new("RACK_2")),
            },
            "SN-SWITCH-001".to_string(),
        );
        let notification = proto::Notification {
            prefix: Some(proto::Path {
                elem: vec![make_path_elem("system-events", &[])],
                ..Default::default()
            }),
            update: vec![
                proto::Update {
                    path: Some(proto::Path {
                        elem: vec![
                            make_path_elem("system-event", &[("event-id", "42")]),
                            make_path_elem("state", &[]),
                            make_path_elem("severity", &[]),
                        ],
                        ..Default::default()
                    }),
                    val: Some(make_typed_value_string("warning")),
                    ..Default::default()
                },
                proto::Update {
                    path: Some(proto::Path {
                        elem: vec![
                            make_path_elem("system-event", &[("event-id", "42")]),
                            make_path_elem("state", &[]),
                            make_path_elem("text", &[]),
                        ],
                        ..Default::default()
                    }),
                    val: Some(make_typed_value_string("Link down detected on swp1")),
                    ..Default::default()
                },
            ],
            ..Default::default()
        };

        assert_eq!(processor.process_notification(&notification), 1);

        let events = sink.events.lock().expect("lock poisoned");
        assert_eq!(events.len(), 1);
        let (context, event) = &events[0];
        assert_eq!(context.switch_id(), Some(switch_id));
        assert_eq!(context.switch_slot_number(), Some(7));
        assert_eq!(context.switch_tray_index(), Some(3));
        assert_eq!(context.rack_id().map(RackId::as_str), Some("RACK_2"));
        let CollectorEvent::Metric(metric) = event else {
            panic!("expected metric event");
        };
        assert_eq!(metric.metric_type, "on_change_row");
        assert_eq!(metric.value, 2.0);
        assert!(
            metric
                .labels
                .iter()
                .any(|(key, value)| key == "instance_id" && value == "42")
        );
    }
}
