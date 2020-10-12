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

	// ThreadID presents the app-specific identifier for grouping notifications.
	ThreadID string `json:"thread-id,omitempty"`

	// Category identifier for custom actions in iOS 8 or newer.
	Category string `json:"category,omitempty"`

	// Content available apps launched in the background or resumed.
	ContentAvailable *int `json:"content-available,omitempty"`

	// The notification service app extension flag. If the value is 1, the system passes the notification to your notification service app extension before delivery.
	MutableContent *int `json:"mutable-content,omitempty"`

	// The identifier of the window brought forward. The value of this key will be populated on the UNNotificationContent object created from the push payload.
	TargetContentId string `json:"target-content-id,omitempty"`
}

// Alert represents aler dictionary.
type Alert struct {
	Title           string   `json:"title,omitempty"`
	Subtitle        string   `json:"subtitle,omitempty"`
	Body            string   `json:"body,omitempty"`
	LaunchImage     string   `json:"launch-image,omitempty"`
	TitleLocKey     string   `json:"title-loc-key,omitempty"`
	TitleLocArgs    []string `json:"title-loc-args,omitempty"`
	SubtitleLocKey  string   `json:"subtitle-loc-key,omitempty"`
	SubtitleLocArgs []string `json:"subtitle-loc-args,omitempty"`
	ActionLocKey    string   `json:"action-loc-key,omitempty"`
	LocKey          string   `json:"loc-key,omitempty"`
	LocArgs         []string `json:"loc-args,omitempty"`
}
