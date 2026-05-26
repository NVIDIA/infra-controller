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

package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// SiteAwaitHandshake waiting for creds request
	SiteAwaitHandshake = "AwaitHandshake"
	// SiteHandshakeComplete creds were given
	SiteHandshakeComplete = "HandshakeComplete"
	// SiteRegistrationComplete site was regsistered, creds no longer
	// available
	SiteRegistrationComplete = "RegistrationComplete"
)

// Site represents one Forge Site
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Site struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SiteSpec `json:"spec,omitempty"`
	// +optional
	Status SiteStatus `json:"status,omitempty"`
}

// SiteSpec represents a site specification
// +k8s:openapi-gen=true
type SiteSpec struct {
	UUID     string `json:"uuid,omitempty"`
	SiteName string `json:"sitename,omitempty"`
	Provider string `json:"provider,omitempty"`
	FCOrg    string `json:"fcorg,omitempty"`
}

// SiteStatus represents the current status of a site
// +k8s:openapi-gen=true
type SiteStatus struct {
	OTP                OTPInfo `json:"otp,omitempty"`
	BootstrapState     string  `json:"bootstrapstate,omitempty"`
	ControlPlaneStatus string  `json:"controlplanestatus,omitempty"`
}

// OTPInfo has a passcode and expiry timestamp
// +k8s:openapi-gen=true
type OTPInfo struct {
	Passcode  string `json:"passcode,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// SiteList is the list of sites
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// no client needed for list as it's been created in above
type SiteList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Site `json:"items"`
}
