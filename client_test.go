package apns

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

func TestSend(t *testing.T) {
	t.Run("send=success", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("apns-collapse-id") != "test-collapse-id" {
				t.Fatalf("expected: %s; got: %s",
					"test-collapse-id",
					req.Header.Get("apns-collapse-id"),
				)
			}
			if req.Header.Get("apns-expiration") != "10" {
				t.Fatalf("expected: %s; got: %s", "10",
					req.Header.Get("apns-expiration"),
				)
			}
			if req.Header.Get("apns-priority") != "5" {
				t.Fatalf("expected: %s; got: %s", "5",
					req.Header.Get("apns-priority"),
				)
			}
			if req.Header.Get("authorization") != "bearer eyJhbGciOiJIUzI1NiIsI" {
				t.Fatalf("expected: %s; got: %s",
					"bearer eyJhbGciOiJIUzI1NiIsI",
					req.Header.Get("authorization"),
				)
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

			rw.Write([]byte(`{"reason": ""}`))
			rw.WriteHeader(http.StatusOK)
		}))
		http2.ConfigureServer(server.Config, nil)
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithAuthorizationToken("bearer eyJhbGciOiJIUzI1NiIsI"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
			WithTimeout(2*time.Second),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.NotificationID != "123e4567-e89b-12d3-a456-42665544000" {
			t.Fatalf("expected apns-id: 123e4567-e89b-12d3-a456-42665544000, but got: %s", resp.NotificationID)
		}
	})

	t.Run("send=fail", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("authorization") != "bearer eyJhbGciOiJIUzI1NiIsI" {
				t.Fatalf("expected: %s; got: %s",
					"bearer eyJhbGciOiJIUzI1NiIsI",
					req.Header.Get("authorization"),
				)
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

			rw.Write([]byte(`{"reason": "BadDeviceToken"}`))
			rw.WriteHeader(http.StatusBadRequest)
		}))
		http2.ConfigureServer(server.Config, nil)
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithAuthorizationToken("bearer eyJhbGciOiJIUzI1NiIsI"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
			WithTimeout(2*time.Second),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
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
		if err != ErrBadDeviceToken {
			t.Fatalf("expected error: %v; got: %v", ErrBadDeviceToken, err)
		}
		if resp.NotificationID != "123e4567-e89b-12d3-a456-42665544000" {
			t.Fatalf("expected apns-id: 123e4567-e89b-12d3-a456-42665544000, but got: %s", resp.NotificationID)
		}
	})
}

func TestSendWithRetry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("apns-collapse-id") != "test-collapse-id" {
				t.Fatalf("expected: %s; got: %s",
					"test-collapse-id",
					req.Header.Get("apns-collapse-id"),
				)
			}
			if req.Header.Get("apns-topic") != "com.app" {
				t.Fatalf("expected: %s; got: %s",
					"com.app",
					req.Header.Get("apns-topic"),
				)
			}
			if req.Header.Get("authorization") != "bearer eyJhbGciOiJIUzI1NiIsI" {
				t.Fatalf("expected: %s; got: %s",
					"bearer eyJhbGciOiJIUzI1NiIsI",
					req.Header.Get("authorization"),
				)
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

			rw.Write([]byte(`{"reason": ""}`))
			rw.WriteHeader(http.StatusOK)
		}))
		http2.ConfigureServer(server.Config, nil)
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithAuthorizationToken("bearer eyJhbGciOiJIUzI1NiIsI"),
			WithBundleID("com.app"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
			WithTimeout(2*time.Second),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := c.SendWithRetry("test-token",
			Payload{
				APS: APS{
					Alert: Alert{
						Title: "hi",
						Body:  "world",
					},
				},
			},
			3,
			WithCollapseID("test-collapse-id"),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.NotificationID != "123e4567-e89b-12d3-a456-42665544000" {
			t.Fatalf("expected apns-id: 123e4567-e89b-12d3-a456-42665544000, but got: %s", resp.NotificationID)
		}
	})

	t.Run("fail", func(t *testing.T) {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("authorization") != "bearer eyJhbGciOiJIUzI1NiIsI" {
				t.Fatalf("expected: %s; got: %s",
					"bearer eyJhbGciOiJIUzI1NiIsI",
					req.Header.Get("authorization"),
				)
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("apns-id", "123e4567-e89b-12d3-a456-42665544000")

			rw.Write([]byte(`{"reason": "error"}`))
			rw.WriteHeader(http.StatusInternalServerError)
		}))
		http2.ConfigureServer(server.Config, nil)
		server.Start()
		defer server.Close()

		c, err := NewClient(
			WithAuthorizationToken("bearer eyJhbGciOiJIUzI1NiIsI"),
			WithEndpoint(server.URL),
			WithMaxIdleConnections(10),
			WithTimeout(2*time.Second),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := c.SendWithRetry("test-token",
			Payload{
				APS: APS{
					Alert: Alert{
						Title: "hi",
						Body:  "world",
					},
				},
			},
			3,
		)
		if err == nil {
			t.Fatal("expected error, got <nil>")
		}
		if resp != nil {
			t.Fatalf("expected <nil> value, but got: %v", resp)
		}
	})
}
