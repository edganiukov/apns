package apns

import "encoding/json"

// Payload repsresents a data structure for APN notification.
type Payload struct {
	APS          APS
	CustomValues map[string]interface{}
}

// MarshalJSON converts Payload structure to the byte array.
// Implements json.Marshaler interface.
func (p Payload) MarshalJSON() ([]byte, error) {
	if p.CustomValues == nil {
		p.CustomValues = make(map[string]interface{})
	}

	p.CustomValues["aps"] = p.APS
	return json.Marshal(p.CustomValues)
}

// APS is Apple's reserved payload.
type APS struct {
	// Alert dictionary.
	Alert Alert `json:"alert,omitempty"`

	// Badge to display on the app icon.
	Badge *int `json:"badge,omitempty"`

	// Sound is the name of a sound file to play as an alert.
	Sound string `json:"sound,omitempty"`

	// Content available apps launched in the background or resumed.
	ContentAvailable bool `json:"content-available,omitempty"`

	// Category identifier for custom actions in iOS 8 or newer.
	Category string `json:"category,omitempty"`

	// ThreadID presents the app-specific identifier for grouping notifications.
	ThreadID string `json:"thread-id,omitempty"`
}

// Alert represents aler dictionary.
type Alert struct {
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	TitleLocKey  string   `json:"title-loc-key,omitempty"`
	TitleLocArgs []string `json:"title-loc-args,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`
}
