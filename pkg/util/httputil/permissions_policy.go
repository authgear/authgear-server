package httputil

import (
	"fmt"
	"strings"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Permissions-Policy#directives
type PermissionsPolicyDirective string

const (
	PermissionsPolicyDirectiveAccelerometer               PermissionsPolicyDirective = "accelerometer"
	PermissionsPolicyDirectiveAmbientLightSensor          PermissionsPolicyDirective = "ambient-light-sensor"
	PermissionsPolicyDirectiveAutoplay                    PermissionsPolicyDirective = "autoplay"
	PermissionsPolicyDirectiveBattery                     PermissionsPolicyDirective = "battery"
	PermissionsPolicyDirectiveBluetooth                   PermissionsPolicyDirective = "bluetooth"
	PermissionsPolicyDirectiveBrowsingTopics              PermissionsPolicyDirective = "browsing-topics"
	PermissionsPolicyDirectiveCamera                      PermissionsPolicyDirective = "camera"
	PermissionsPolicyDirectiveDisplayCapture              PermissionsPolicyDirective = "display-capture"
	PermissionsPolicyDirectiveDocumentDomain              PermissionsPolicyDirective = "document-domain"
	PermissionsPolicyDirectiveEncryptedMedia              PermissionsPolicyDirective = "encrypted-media"
	PermissionsPolicyDirectiveExecutionWhileNotRendered   PermissionsPolicyDirective = "execution-while-not-rendered"
	PermissionsPolicyDirectiveExecutionWhileOutOfViewport PermissionsPolicyDirective = "execution-while-out-of-viewport"
	PermissionsPolicyDirectiveFullscreen                  PermissionsPolicyDirective = "fullscreen"
	PermissionsPolicyDirectiveGamepad                     PermissionsPolicyDirective = "gamepad"
	PermissionsPolicyDirectiveGeolocation                 PermissionsPolicyDirective = "geolocation"
	PermissionsPolicyDirectiveGyroscope                   PermissionsPolicyDirective = "gyroscope"
	PermissionsPolicyDirectiveHid                         PermissionsPolicyDirective = "hid"
	PermissionsPolicyDirectiveIdentityCredentialsGet      PermissionsPolicyDirective = "identity-credentials-get"
	PermissionsPolicyDirectiveIdleDetection               PermissionsPolicyDirective = "idle-detection"
	PermissionsPolicyDirectiveLocalFonts                  PermissionsPolicyDirective = "local-fonts"
	PermissionsPolicyDirectiveMagnetometer                PermissionsPolicyDirective = "magnetometer"
	PermissionsPolicyDirectiveMicrophone                  PermissionsPolicyDirective = "microphone"
	PermissionsPolicyDirectiveMidi                        PermissionsPolicyDirective = "midi"
	PermissionsPolicyDirectiveOtpCredentials              PermissionsPolicyDirective = "otp-credentials"
	PermissionsPolicyDirectivePayment                     PermissionsPolicyDirective = "payment"
	PermissionsPolicyDirectivePictureInPicture            PermissionsPolicyDirective = "picture-in-picture"
	PermissionsPolicyDirectivePublickeyCredentialsCreate  PermissionsPolicyDirective = "publickey-credentials-create"
	PermissionsPolicyDirectivePublickeyCredentialsGet     PermissionsPolicyDirective = "publickey-credentials-get"
	PermissionsPolicyDirectiveScreenWakeLock              PermissionsPolicyDirective = "screen-wake-lock"
	PermissionsPolicyDirectiveSerial                      PermissionsPolicyDirective = "serial"
	PermissionsPolicyDirectiveSpeakerSelection            PermissionsPolicyDirective = "speaker-selection"
	PermissionsPolicyDirectiveStorageAccess               PermissionsPolicyDirective = "storage-access"
	PermissionsPolicyDirectiveUsb                         PermissionsPolicyDirective = "usb"
	PermissionsPolicyDirectiveWebShare                    PermissionsPolicyDirective = "web-share"
	PermissionsPolicyDirectiveWindowManagement            PermissionsPolicyDirective = "window-management"
	PermissionsPolicyDirectiveXrSpatialTracking           PermissionsPolicyDirective = "xr-spatial-tracking"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Permissions-Policy#allowlist
type PermissionsPolicyAllowlist string

const (
	PermissionsPolicyAllowlistAll  PermissionsPolicyAllowlist = "*"
	PermissionsPolicyAllowlistNone PermissionsPolicyAllowlist = "()"
	PermissionsPolicyAllowlistSelf PermissionsPolicyAllowlist = "(self)"
	PermissionsPolicyAllowlistSrc  PermissionsPolicyAllowlist = "(src)"
)

type PermissionsPolicyPolicy struct {
	Directive PermissionsPolicyDirective
	Allowlist PermissionsPolicyAllowlist
}

/*
 * Enabled features:
 * - autoplay=*
 * - execution-while-not-rendered=*
 * - execution-while-out-of-viewport=*
 * - fullscreen=*
 * - publickey-credentials-create=(self) (for WebAuthn)
 * - publickey-credentials-get=(self) (for WebAuthn)
 */
var DefaultPermissionsPolicy = []PermissionsPolicyPolicy{
	{PermissionsPolicyDirectiveAccelerometer, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveAmbientLightSensor, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveAutoplay, PermissionsPolicyAllowlistAll},
	{PermissionsPolicyDirectiveBattery, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveBluetooth, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveBrowsingTopics, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveCamera, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveDisplayCapture, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveDocumentDomain, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveEncryptedMedia, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveExecutionWhileNotRendered, PermissionsPolicyAllowlistAll},
	{PermissionsPolicyDirectiveExecutionWhileOutOfViewport, PermissionsPolicyAllowlistAll},
	{PermissionsPolicyDirectiveFullscreen, PermissionsPolicyAllowlistAll},
	{PermissionsPolicyDirectiveGamepad, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveGeolocation, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveGyroscope, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveHid, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveIdentityCredentialsGet, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveIdleDetection, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveLocalFonts, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveMagnetometer, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveMicrophone, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveMidi, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveOtpCredentials, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectivePayment, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectivePictureInPicture, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectivePublickeyCredentialsCreate, PermissionsPolicyAllowlistSelf},
	{PermissionsPolicyDirectivePublickeyCredentialsGet, PermissionsPolicyAllowlistSelf},
	{PermissionsPolicyDirectiveScreenWakeLock, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveSerial, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveSpeakerSelection, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveStorageAccess, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveUsb, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveWebShare, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveWindowManagement, PermissionsPolicyAllowlistNone},
	{PermissionsPolicyDirectiveXrSpatialTracking, PermissionsPolicyAllowlistNone},
}

type HTTPPermissionsPolicy []PermissionsPolicyPolicy

func (p HTTPPermissionsPolicy) String() string {
	var parts []string
	for _, policy := range p {
		parts = append(parts, fmt.Sprintf("%s=%s", policy.Directive, policy.Allowlist))
	}
	return strings.Join(parts, ", ")
}
