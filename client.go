package apns

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// APN service endpoint URLs.
const (
	SandboxGateway    = "https://api.sandbox.push.apple.com"
	ProductionGateway = "https://api.push.apple.com"
)

// JWT represents data for JWT token generation.
type JWT struct {
	PrivateKey *ecdsa.PrivateKey
	Issuer     string
	KeyID      string
}

// Client represents the Apple Push Notification Service that you send notifications to.
type Client struct {
	http     *http.Client
	endpoint string
	jwt      *JWT

	mtx      sync.RWMutex
	sendOpts map[string]SendOption
}

// NewClient creates new AONS client based on defined Options.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		http: &http.Client{
			Transport: &http.Transport{},
		},
		endpoint: ProductionGateway,
		sendOpts: make(map[string]SendOption),
	}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Send sends Notification to the APN service.
func (c *Client) Send(ctx context.Context, token string, p Payload, opts ...SendOption) (*Response, error) {
	req, err := c.prepareRequest(ctx, token, p, opts...)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, req)
}

func (c *Client) prepareRequest(ctx context.Context, tok string, p Payload, opts ...SendOption) (*http.Request, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/3/device/%s", c.endpoint, tok),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	c.mtx.RLock()
	// apply send options
	for _, o := range c.sendOpts {
		o(req.Header)
	}
	c.mtx.RUnlock()

	for _, o := range opts {
		o(req.Header)
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*Response, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, connError(err.Error())
	}
	defer resp.Body.Close()

	response := new(Response)
	response.NotificationID = resp.Header.Get("apns-id")

	switch resp.StatusCode {
	case http.StatusOK:
		return response, nil
	case http.StatusForbidden:
		if c.jwt != nil {
			token, err := c.issueToken()
			if err != nil {
				return nil, err
			}
			req.Header.Set("authorization", fmt.Sprintf("bearer %s", token))
			return c.do(ctx, req)
		}
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		return nil, serverError(fmt.Sprintf("%d error: %s", resp.StatusCode, resp.Status))
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}
	return response, response.Error
}

func (c *Client) issueToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": c.jwt.Issuer,
		"iat": time.Now().Unix(),
	})
	token.Header["kid"] = c.jwt.KeyID

	t, err := token.SignedString(c.jwt.PrivateKey)
	if err != nil {
		return "", err
	}

	c.mtx.Lock()
	c.sendOpts["authorization"] = func(h http.Header) {
		h.Set("authorization", fmt.Sprintf("bearer %s", t))
	}
	c.mtx.Unlock()

	return t, nil
}
