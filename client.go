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

var (
	defaultTokenRenewInterval    = 10 * time.Minute
	defaultTokenValidityInterval = time.Hour
)

// JWTConfig represents configuration to generate JWT.
type JWTConfig struct {
	PrivateKey *ecdsa.PrivateKey
	Issuer     string
	KeyID      string
}

// Client represents the Apple Push Notification Service that you send notifications to.
type Client struct {
	http      *http.Client
	endpoint  string
	jwtConfig *JWTConfig

	mtx      sync.RWMutex
	sendOpts map[string]SendOption
}

// NewClient creates new AONS client based on defined Options.
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
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

	if c.jwtConfig != nil {
		go c.renewToken(ctx, defaultTokenRenewInterval)
	}

	return c, nil
}

// Send sends Notification to the APN service.
func (c *Client) Send(ctx context.Context, deviceToken string, p Payload, opts ...SendOption) (*Response, error) {
	req, err := c.newRequest(ctx, deviceToken, p, opts...)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, req)
}

func (c *Client) newRequest(ctx context.Context, token string, p Payload, opts ...SendOption) (*http.Request, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/3/device/%s", c.endpoint, token),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	c.mtx.RLock()
	// If JWT is used, sendOpts sets `Authorization` header.
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
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		return nil, serverError(fmt.Sprintf("%d error: %s", resp.StatusCode, resp.Status))
	default:
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return nil, err
		}
		return response, response.Error
	}
}

func (c *Client) renewToken(ctx context.Context, renewInterval time.Duration) {
	tick := time.NewTicker(renewInterval)
	for {
		select {
		case <-tick.C:
			token, err := c.issueToken()
			if err != nil {

			}

			c.mtx.Lock()
			c.sendOpts["authorization"] = WithAuthorizationToken(token)
			c.mtx.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) issueToken() (string, error) {
	tNow := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.RegisteredClaims{
		Issuer:    c.jwtConfig.Issuer,
		IssuedAt:  jwt.NewNumericDate(tNow),
		ExpiresAt: jwt.NewNumericDate(tNow.Add(defaultTokenValidityInterval)),
	})
	token.Header["kid"] = c.jwtConfig.KeyID

	t, err := token.SignedString(c.jwtConfig.PrivateKey)
	if err != nil {
		return "", err
	}

	return t, nil
}
