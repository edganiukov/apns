package apns

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strconv"
	"time"
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
// Certificates should be used with BundleID, that is possible to set by
// `WithBundleID` option.
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

// WithTimeout sets HTTP Client timeout.
func WithTimeout(t time.Duration) ClientOption {
	return func(c *Client) error {
		c.http.Timeout = t
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

// WithAuthorizationToken sets the token that authorizes APNs to send push
// notifications for the specified topics. The token is in Base64URL-encoded
// JWT format, specified as `bearer <provider token>`.
func WithAuthorizationToken(token string) ClientOption {
	return func(c *Client) error {
		if token == "" {
			return errors.New("invalid authorization token")
		}
		c.sendOpts = append(c.sendOpts, func(h http.Header) {
			h.Add("authorization", token)
		})
		return nil
	}
}

// WithBundleID sets HTTP2 header `apns-topic` with is bundle ID of an app.
// The certificate you create in your developer account must include
// the capability for this topic. If your certificate includes multiple topics,
// you must specify a value for this header. If you omit this request header
// and your APNs certificate does not specify multiple topics, the APNs server
// uses the certificate’s Subject as the default topic. If you are using
// a provider token instead of a certificate, you must specify a value
// for this request header. The topic you provide should be provisioned for
// the your team named in your developer account.
func WithBundleID(bundleID string) ClientOption {
	return func(c *Client) error {
		if bundleID == "" {
			return errors.New("invalid bundle ID")
		}
		c.sendOpts = append(c.sendOpts, func(h http.Header) {
			h.Add("apns-topic", bundleID)
		})
		return nil
	}
}

// SendOption allows to set custom Headers for each notification, such apns-id,
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
