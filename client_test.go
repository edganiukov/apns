package apns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testPrivateKey = []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgddHKwoJxaoSW6UQt
DIPli9xODjt+6DWVjgELdB8NOHKhRANCAATzgguZzRtO6prpwKRBsCA06seZMvDd
y6bx7sCtRZz9kvZjSLox5kEqVaEZgEgDwUIKY29Wl0weel+1hChax3OS
-----END PRIVATE KEY-----`)

func TestSend(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(
			func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, req.Header.Get("apns-collapse-id"), "test-collapse-id")
				assert.Equal(t, req.Header.Get("apns-expiration"), "10")
				assert.Equal(t, req.Header.Get("apns-priority"), "5")

				rw.Header().Set("Content-Type", "application/json")
				rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{"reason": ""}`))
			},
		))
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithJWT(testPrivateKey, "key_id", "issuer"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
		)
		assert.NoError(t, err)
		assert.NotNil(t, c)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		resp, err := c.Send(ctx, "test-token",
			Payload{
				APS: APS{
					Alert: Alert{
						Title: "hi",
						Body:  "world",
					},
				},
			},
			WithExpiration(10),
			WithCollapseID("test-collapse-id"),
			WithPriority(5),
		)
		assert.NoError(t, err)
		assert.Equal(t, resp.NotificationID, "123e4567-e89b-12d3-a456-42665544000")
	})

	t.Run("invalid device token", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"reason": "BadDeviceToken"}`))
		}))
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithJWT(testPrivateKey, "key_id", "issuer"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
		)
		assert.NoError(t, err)

		resp, err := c.Send(context.Background(), "",
			Payload{
				APS: APS{
					Alert: Alert{
						Title: "hi",
						Body:  "world",
					},
				},
			},
		)
		assert.Equal(t, err, ErrBadDeviceToken)
		assert.Equal(t, resp.NotificationID, "123e4567-e89b-12d3-a456-42665544000")
	})
}
