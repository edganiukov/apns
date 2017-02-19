package apns

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/http2"
)

func TestSend(t *testing.T) {
	pKey, err := ioutil.ReadFile("testdata/key.pem")
	assert.NoError(t, err)

	t.Run("send=success", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Header.Get("apns-collapse-id"), "test-collapse-id")
			assert.Equal(t, req.Header.Get("apns-expiration"), "10")
			assert.Equal(t, req.Header.Get("apns-priority"), "5")

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(`{"reason": ""}`))
		}))
		http2.ConfigureServer(server.Config, nil)
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithJWT(pKey, "key_id", "issuer"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
			WithTimeout(2*time.Second),
		)

		assert.NoError(t, err)

		resp, err := c.Send("test-token",
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

	t.Run("send=fail", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"reason": "BadDeviceToken"}`))
		}))
		http2.ConfigureServer(server.Config, nil)
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithJWT(pKey, "key_id", "issuer"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
			WithTimeout(2*time.Second),
		)
		assert.NoError(t, err)

		resp, err := c.Send("",
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
