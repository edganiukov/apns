# apns
[![GoDoc](https://pkg.go.dev/badge/github.com/edganiukov/apns)](https://pkg.go.dev/github.com/edganiukov/apns)

Golang client library for Apple Push Notification service via HTTP2.
More information on [Apple Push Notification Service](https://developer.apple.com/library/content/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/APNSOverview.html)

### Getting Started
-------------------
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
		apns.WithJWT(data, "key_id", "team_id"),
		apns.WithAppID("app_id"),
		apns.WithMaxIdleConnections(10),
		apns.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := c.Send(ctx, "<device token>",
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
		log.Fatal(err)
	}

    /* ... */
}
```
In case, if you want to use TLS certificate instead of JWT tokens, then should
use `apns.WithCertificate` and `apns.WithAppID` `ClientOption` to specify
certificate and app ID, that are needed to send push notifications.
