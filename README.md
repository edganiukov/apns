# apns
[![GoDoc](https://godoc.org/github.com/edganiukov/fcm?status.svg)](https://godoc.org/github.com/edganiukov/apns)

Golang client library for Apple Push Notification service via HTTP2. More information on [Apple Push Notification Service](https://developer.apple.com/library/content/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/APNSOverview.html)

### Getting Started
-------------------
To install apns, use `go get`:

```bash
go get github.com/edganiukov/apns
```
or `govendor`:

```bash
govendor fetch github.com/edganiukov/apns
```
or other tool for vendoring.

### Sample Usage
----------------
Here is a simple example illustrating how to use APNS library:
```go
package main

import (
	"github.com/edganiukov/apns"
)

func main() {
	data, err := ioutil.ReadFile("private_key.pem")
	if err != nil {
		log.Fatal(err)
	}

	c, err := apns.NewClient(
		apns.WithJWT(data, "key_identifier", "team_identifier"),
		apns.WithBundleID("bundle_idetifier"),
		apns.WithMaxIdleConnections(10),
		apns.WithTimeout(5*time.Second),
	)
	if err != nil {
		/* ... */
	}
	resp, err := c.Send("<device token>",
		apns.Payload{
			APS: apns.APS{
				Alert: apns.Alert{
					Title: "Test Push",
					Body:  "Hi world",
				},
			},
		},
		apns.WithExpiration(10),
		apns.WithCollapseID("test-collapse-id"),
		apns.WithPriority(5),
	)

	if err != nil {
		/* ... */
	}
}
```
In case, if you wanna use TLS certificate instead of JWT Token, then please use `apns.WithCertificate` and `apns.WithBundleID` CallOptions to specify certificate and bundle ID, that are needed to send pushes.
