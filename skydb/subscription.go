// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package skydb

import "errors"

// ErrSubscriptionNotFound is returned from GetSubscription or
// DeleteSubscription when the specific subscription cannot be found.
var ErrSubscriptionNotFound = errors.New("skydb: Subscription ID not found")

// Subscription represents a device's subscription of notification
// triggered by changes of results from a query.
type Subscription struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	DeviceID         string            `json:"device_id"`
	NotificationInfo *NotificationInfo `json:"notification_info,omitempty"`
	Query            Query             `json:"query"`
}

// NotificationInfo describes how server should send a notification
// to a target devices via a push service. Currently only APS is supported.
type NotificationInfo struct {
	APS APSSetting `json:"aps,omitempty"`
}

// APSSetting describes how server should send a notification to a
// targeted device via Apple Push Service.
type APSSetting struct {
	Alert                      *AppleAlert `json:"alert,omitempty"`
	SoundName                  string      `json:"sound,omitempty"`
	ShouldBadge                bool        `json:"should-badge,omitempty"`
	ShouldSendContentAvailable bool        `json:"should-send-content-available,omitempty"`
}

// AppleAlert describes how a remote notification behaves and shows
// itself when received.
//
// It is a subset of attributes defined in Apple's "Local and Remote
// Notification Programming Guide". Please follow the following link
// for detailed description of the attributes.
//	https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/ApplePushService.html#//apple_ref/doc/uid/TP40008194-CH100-SW20
type AppleAlert struct {
	Body                  string   `json:"body,omitempty"`
	LocalizationKey       string   `json:"loc-key,omitempty"`
	LocalizationArgs      []string `json:"loc-args,omitempty"`
	LaunchImage           string   `json:"launch-image,omitempty"`
	ActionLocalizationKey string   `json:"action-loc-key,omitempty"`
}
