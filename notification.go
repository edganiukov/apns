package apns

import "encoding/json"

// Payload repsresents a data structure for APN notification.
type Payload struct {
	APS          APS
	CustomValues map[string]any
}

// MarshalJSON converts Payload structure to the byte array.
// Implements json.Marshaler interface.
func (p Payload) MarshalJSON() ([]byte, error) {
	if p.CustomValues == nil {
		p.CustomValues = make(map[string]any)
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

	// The notification service app extension flag. If the value is 1, the system passes the notification to your
	// notification service app extension before delivery.
	MutableContent *int `json:"mutable-content,omitempty"`

	// The identifier of the window brought forward. The value of this key will be populated on the
	// UN NotificationContent object created from the push payload.
	TargetContentID string `json:"target-content-id,omitempty"`

	// The importance and delivery timing of a notification. The string values “passive”, “active”, “time-sensitive”,
	// or “critical” correspond to the UNNotificationInterruptionLevel enumeration cases.
	InteraptionLevel string `json:"interruption-level,omitempty"`

	// The relevance score, a number between 0 and 1, that the system uses to sort the notifications from your app.
	// The highest score gets featured in the notification summary.
	RelevanceScore *int `json:"relevance-score,omitempty"`

	// The criteria the system evaluates to determine if it displays the notification in the current Focus.
	FilterCriteria string `json:"filter-criteria,omitempty"`

	// The UNIX timestamp that represents the date at which a Live Activity becomes stale, or out of date.
	StaleDate *int `json:"stale-date,omitempty"`

	// The updated or final content for a Live Activity. The content of this dictionary must match the data you
	// describe with your custom ActivityAttributes implementation.
	ContentStale map[string]string `json:"content-state,omitempty"`

	// The UNIX timestamp that marks the time when you send the remote notification that updates or ends a Live
	// Activity.
	Timestamp *int `json:"timestamp,omitempty"`

	// The string that describes whether you update or end an ongoing Live Activity with the remote push notification.
	// To update the Live Activity, use update. To end the Live Activity, use end.
	Events string `json:"events,omitempty"`
}

// Alert represents aler dictionary.
type Alert struct {
	// The title of the notification. Apple Watch displays this string in the short look notification interface.
	Title string `json:"title,omitempty"`
	// Additional information that explains the purpose of the notification.
	Subtitle string `json:"subtitle,omitempty"`
	// The content of the alert message.
	Body string `json:"body,omitempty"`
	// The name of the launch image file to display. If the user chooses to launch your app, the contents of the
	// specified image or storyboard file are displayed instead of your app’s normal launch image.
	LaunchImage string `json:"launch-image,omitempty"`
	// The key for a localized title string. Specify this key instead of the title key to retrieve the title from your
	// app’s Localizable.strings files. The value must contain the name of a key in your strings file.
	TitleLocKey string `json:"title-loc-key,omitempty"`
	// An array of strings containing replacement values for variables in your title string. Each %@ character in the
	// string specified by the title-loc-key is replaced by a value from this array. The first item in the array
	// replaces the first instance of the %@ character in the string, the second item replaces the second instance,
	// and so on.
	TitleLocArgs []string `json:"title-loc-args,omitempty"`
	// The key for a localized subtitle string. Use this key, instead of the subtitle key, to retrieve the subtitle from
	// your app’s Localizable.strings file. The value must contain the name of a key in your strings file.
	SubtitleLocKey string `json:"subtitle-loc-key,omitempty"`
	// An array of strings containing replacement values for variables in your title string. Each %@ character in the
	// string specified by subtitle-loc-key is replaced by a value from this array. The first item in the array replaces
	// the first instance of the %@ character in the string, the second item replaces the second instance, and so on.
	SubtitleLocArgs []string `json:"subtitle-loc-args,omitempty"`
	// The key for a localized message string. Use this key, instead of the body key, to retrieve the message text from
	// your app’s Localizable.strings file. The value must contain the name of a key in your strings file.
	LocKey string `json:"loc-key,omitempty"`
	// An array of strings containing replacement values for variables in your message text. Each %@ character in the
	// string specified by loc-key is replaced by a value from this array. The first item in the array replaces the
	// first instance of the %@ character in the string, the second item replaces the second instance, and so on.
	LocArgs []string `json:"loc-args,omitempty"`
}

// Pointer returns a pointer to a provided value.
func Pointer[T any](v T) *T {
	return &v
}
