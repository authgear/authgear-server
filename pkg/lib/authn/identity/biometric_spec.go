package identity

type BiometricSpec struct {
	KeyID      string                 `json:"key_id,omitempty"`
	Key        string                 `json:"key,omitempty"`
	DeviceInfo map[string]interface{} `json:"device_info,omitempty"`
}
