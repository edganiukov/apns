package apns

import (
	"encoding/json"
	"errors"
)

// Possible error codes included in the reason key of a response’s JSON payload.
var (
	ErrBadCollapseID               = errors.New("collapse identifier exceeds the maximum allowed size")
	ErrBadDeviceToken              = errors.New("specified device token was bad")
	ErrBadExpirationDate           = errors.New("apns-expiration value is bad")
	ErrBadMessageID                = errors.New("apns-id value is bad")
	ErrBadPriority                 = errors.New("apns-priority value is bad")
	ErrBadTopic                    = errors.New("apns-topic was invalid")
	ErrDeviceTokenNotForTopic      = errors.New("device token does not match the specified topic")
	ErrDuplicateHeaders            = errors.New("one or more headers were repeated")
	ErrIdleTimeout                 = connError("idle time out")
	ErrMissingDeviceToken          = errors.New("device token is not specified in the request path")
	ErrMissingTopic                = errors.New("apns-topic header of the request was not specified and was required")
	ErrPayloadEmpty                = errors.New("message payload was empty")
	ErrTopicDisallowed             = errors.New("pushing to this topic is not allowed")
	ErrBadCertificate              = errors.New("certificate was bad")
	ErrBadCertificateEnvironment   = errors.New("client certificate was for the wrong environment")
	ErrExpiredProviderToken        = errors.New("provider token is stale and a new token should be generated")
	ErrForbidden                   = errors.New("specified action is not allowed")
	ErrInvalidProviderToken        = errors.New("provider token is not valid or the token signature could not be verified")
	ErrMissingProviderToken        = errors.New("no provider certificate was used to connect to APNs and Authorization header was missing")
	ErrBadPath                     = errors.New("request contained a bad :path value")
	ErrMethodNotAllowed            = errors.New("specified method was not POST")
	ErrUnregistered                = errors.New("device token is inactive for the specified topic")
	ErrPayloadTooLarge             = errors.New("message payload was too large")
	ErrTooManyProviderTokenUpdates = errors.New("provider token is being updated too often")
	ErrTooManyRequests             = errors.New("too many requests were made consecutively to the same device token")
	ErrInternalServerError         = serverError("an internal server error occurred")
	ErrServiceUnavailable          = serverError("service is unavailable")
	ErrShutdown                    = serverError("the server is shutting down")
)

var errorsMapping = map[string]error{
	"BadCollapseID":               ErrBadCollapseID,
	"BadDeviceToken":              ErrBadDeviceToken,
	"BadExpirationDate":           ErrBadExpirationDate,
	"BadMessageId":                ErrBadMessageID,
	"BadPriority":                 ErrBadPriority,
	"BadTopic":                    ErrBadTopic,
	"DeviceTokenNotForTopic":      ErrDeviceTokenNotForTopic,
	"DuplicateHeaders":            ErrDuplicateHeaders,
	"IdleTimeout":                 ErrIdleTimeout,
	"MissingDeviceToken":          ErrMissingDeviceToken,
	"MissingTopic":                ErrMissingTopic,
	"PayloadEmpty":                ErrPayloadEmpty,
	"TopicDisallowed":             ErrTopicDisallowed,
	"BadCertificate":              ErrBadCertificate,
	"BadCertificateEnvironment":   ErrBadCertificateEnvironment,
	"ExpiredProviderToken":        ErrExpiredProviderToken,
	"Forbidden":                   ErrForbidden,
	"InvalidProviderToken":        ErrInvalidProviderToken,
	"MissingProviderToken":        ErrMissingProviderToken,
	"BadPath":                     ErrBadPath,
	"MethodNotAllowed":            ErrMethodNotAllowed,
	"Unregistered":                ErrUnregistered,
	"PayloadTooLarge":             ErrPayloadTooLarge,
	"TooManyProviderTokenUpdates": ErrTooManyProviderTokenUpdates,
	"TooManyRequests":             ErrTooManyRequests,
	"InternalServerError":         ErrInternalServerError,
	"ServiceUnavailable":          ErrServiceUnavailable,
	"Shutdown":                    ErrShutdown,
}

type connError string

func (e connError) Error() string {
	return string(e)
}

func (e connError) Temporary() bool {
	return true
}

func (e connError) Timeout() bool {
	return true
}

type serverError string

func (e serverError) Error() string {
	return string(e)
}

func (e serverError) Temporary() bool {
	return true
}

func (e serverError) Timeout() bool {
	return false
}

// Response represents response object from APN service.
type Response struct {
	NotificationID string
	Timestamp      int64
	Error          error
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Response) UnmarshalJSON(data []byte) error {
	var rawResp struct {
		Reason    string `json:"reason"`
		Timestamp int64  `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &rawResp); err != nil {
		return err
	}
	if rawResp.Reason != "" {
		if err, ok := errorsMapping[rawResp.Reason]; ok {
			r.Error = err
		} else {
			r.Error = errors.New("unknown error")
		}
	}
	r.Timestamp = rawResp.Timestamp
	return nil
}
