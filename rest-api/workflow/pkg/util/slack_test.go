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

package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlackClient_sendHttpRequest(t *testing.T) {
	type fields struct {
		WebHookUrl string
		TimeOut    time.Duration
	}
	type args struct {
		slackRequest SlackMessage
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test sending http request with Slack message",
			fields: fields{
				WebHookUrl: testServer.URL,
				TimeOut:    5 * time.Second,
			},
			args: args{
				slackRequest: SlackMessage{
					IconEmoji: ":ghost:",
					Text:      "This is a test message",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := SlackClient{
				webHookUrl: tt.fields.WebHookUrl,
				timeOut:    tt.fields.TimeOut,
			}

			err := sc.sendHttpRequest(tt.args.slackRequest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
