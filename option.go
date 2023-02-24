package apns

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ClientOption defines athe APNS Client option.
type ClientOption func(c *Client) error

// WithHTTPClient sets custom HTTP Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		c.http = httpClient
		return nil
	}
}

// WithEndpoint specifies custom APN endpoint. Useful for test propose.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) error {
		c.endpoint = endpoint
		return nil
	}
}

// WithCertificate is Option to configure TLS certificates for HTTP connection.
// Certificates should be used with app ID, that is possible to set by
// [WithAppID] option.
func WithCertificate(crt tls.Certificate) ClientOption {
	return func(c *Client) error {
		config := &tls.Config{
			Certificates: []tls.Certificate{crt},
		}
		config.BuildNameToCertificate()
		c.http.Transport.(*http.Transport).TLSClientConfig = config
		return nil
	}
}

// WithMaxIdleConnections sets maximum number of the idle HTTP connection
// that can be reused in order do not create new connection.
func WithMaxIdleConnections(maxIdleConn int) ClientOption {
	return func(c *Client) error {
		if maxIdleConn < 1 {
			return errors.New("invalid MaxIdleConnsPerHost")
		}
		c.http.Transport.(*http.Transport).MaxIdleConnsPerHost = maxIdleConn
		return nil
	}
}

// WithJWT sets the JWT config that is used to generate a JWT token to authorize against APNS to send push
// notifications for the specified topics. The token is in Base64URL-encoded JWT format, specified as
// `bearer <provider token>`.
func WithJWT(privateKey []byte, keyID string, teamID string) ClientOption {
	return func(c *Client) error {
		key, err := parsePrivateKey(privateKey)
		if err != nil {
			return err
		}
		c.jwtConfig = &JWTConfig{
			PrivateKey: key,
			KeyID:      keyID,
			Issuer:     teamID,
		}

		token, err := c.issueToken()
		if err != nil {
			return err
		}

		c.sendOpts["authorization"] = WithAuthorizationToken(token)
		return nil
	}
}

// WithBundleID sets HTTP2 header `apns-topic` with is bundle ID of an app.
//
// Deprecated: use [WithAppID]
func WithBundleID(bundleID string) ClientOption {
	return func(c *Client) error {
		if bundleID == "" {
			return errors.New("invalid bundle ID")
		}

		c.sendOpts["apns-topic"] = func(h http.Header) {
			h.Set("apns-topic", bundleID)
		}

		return nil
	}
}

// WithAppID sets HTTP2 header `apns-topic` with an application ID.
// An App ID identifies your app in a provisioning profile. It is a two-part string used to identify one or more apps
// from a single development team. There are two types of App IDs: an explicit App ID, used for a single app, and a
// wildcard App ID, used for a set of apps.
func WithAppID(appID string) ClientOption {
	return func(c *Client) error {
		if appID == "" {
			return errors.New("invalid application ID")
		}

		c.sendOpts["apns-topic"] = func(h http.Header) {
			h.Set("apns-topic", appID)
		}

		return nil
	}
}

// SendOption allows to set custom Headers for each notification, such as apns-id,
// expiration time, priority, etc.
type SendOption func(h http.Header)

// WithNotificationID sets a  canonical UUID that identifies the notification.
// If there is an error sending the notification, APNs uses this value
// to identify the notification to your server. The canonical form is
// 32 lowercase hexadecimal digits, displayed in five groups separated by
// hyphens in the form 8-4-4-4-12. If you omit this option,
// a new UUID is created by APNs and returned in the response.
func WithNotificationID(id string) SendOption {
	return func(h http.Header) {
		h.Set("apns-id", id)
	}
}

// WithExpiration sets a headers, that identifies the date when the notification
// is no longer valid and can be discarded. If this value is nonzero,
// APNs stores the notification and tries to deliver it at least once, repeating
// the attempt as needed if it is unable to deliver the notification the first time.
// If the value is 0, APNs treats the notification as if it expires immediately
// and does not store the notification or attempt to redeliver it.
func WithExpiration(timeExpr int) SendOption {
	return func(h http.Header) {
		h.Set("apns-expiration", strconv.Itoa(timeExpr))
	}
}

// WithPriority specifies the  priority of the notification.
// Specify one of the following values:
// * 10 - Send the push message immediately. Notifications with this priority
// must trigger an alert, sound, or badge on the target device.
// It is an error to use this priority for a push notification that contains
// only the content-available key.
// * 5 - Send the push message at a time that takes into account power
// considerations for the device. Notifications with this priority might be grouped
// and delivered in bursts. They are throttled, and in some cases are not delivered.
func WithPriority(priority int) SendOption {
	return func(h http.Header) {
		h.Set("apns-priority", strconv.Itoa(priority))
	}
}

// WithCollapseID sets commond idetifier for Multiple notifications,
// which will be displayed to the user as a single notification.
// The value of this key must not exceed 64 bytes.
func WithCollapseID(id string) SendOption {
	return func(h http.Header) {
		h.Set("apns-collapse-id", id)
	}
}

// WithPushType sets a value of the `apns-push-type` header that accurately reflect the contents of your notification’s
// payload. If there’s a mismatch, or if the header is missing on required systems, APNs may return an error, delay the
// delivery of the notification, or drop it altogether.
// Required for watchOS 6 and later; recommended for macOS, iOS, tvOS, and iPadOS)

// The apns-push-type header field has the following valid values. The descriptions below describe when and how to use
// these values.
//
//   - alert
//     Use the alert push type for notifications that trigger a user interaction—for example, an alert, badge, or sound.
//
//   - background
//     Use the background push type for notifications that deliver content in the background, and don’t trigger any user
//     interactions.
//
//   - location
//     Use the location push type for notifications that request a user’s location.
//
//   - voip
//     Use the voip push type for notifications that provide information about an incoming Voice-over-IP (VoIP) call.
//
//   - complication
//     Use the complication push type for notifications that contain update information for a watchOS app’s
//     complications.
//
//   - fileprovider
//     Use the fileprovider push type to signal changes to a File Provider extension.
//
//   - mdm
//     Use the mdm push type for notifications that tell managed devices to contact the MDM server.
//
//   - liveactivity
//     Use the liveactivity push type to send a remote push notification that updates or ends an ongoing Live Activity.
func WithPushType(t string) SendOption {
	return func(h http.Header) {
		h.Set("apns-push-type", t)
	}
}

// WithAuthorizationToken sets `Authorization` header with a bearer token.
func WithAuthorizationToken(t string) SendOption {
	return func(h http.Header) {
		h.Set("authorization", fmt.Sprintf("bearer %s", t))
	}
}

func parsePrivateKey(key []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("not PEM encoded key")
	}
	pKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := pKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("not ECDSA private key")
	}
	return k, nil
}
