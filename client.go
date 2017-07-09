package apns

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"golang.org/x/net/http2"
)

// APN service endpoint URLs.
const (
	DevelopmentGateway = "https://api.development.push.apple.com"
	ProductionGateway  = "https://api.push.apple.com"
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
	jwt      *JWT
	sendOpts map[string]SendOption
	endpoint string
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
	if err := http2.ConfigureTransport(c.http.Transport.(*http.Transport)); err != nil {
		return nil, err
	}

	return c, nil
}

// Send sends Notification to the APN service.
func (c *Client) Send(token string, p Payload, opts ...SendOption) (*Response, error) {
	req, err := c.prepareRequest(token, p, opts...)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) prepareRequest(token string, p Payload, opts ...SendOption) (*http.Request, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/3/device/%s", c.endpoint, token),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// apply send options
	for _, o := range c.sendOpts {
		o(req.Header)
	}
	for _, o := range opts {
		o(req.Header)
	}
	return req, nil
}

func (c *Client) do(req *http.Request) (*Response, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, connError(err.Error())
	}
	defer resp.Body.Close()

	response := new(Response)
	response.NotificationID = resp.Header.Get("apns-id")

	if resp.StatusCode == http.StatusOK {
		return response, nil
	}
	if resp.StatusCode == http.StatusForbidden {
		if c.jwt == nil {
			if resp.StatusCode >= http.StatusInternalServerError {
				return nil, serverError(fmt.Sprintf("%d error: %s", resp.StatusCode, resp.Status))
			}
			return nil, fmt.Errorf("%d error: %s", resp.StatusCode, resp.Status)
		}

		// in case of using JWT
		token, err := c.issueToken()
		if err != nil {
			return nil, err
		}
		req.Header.Set("authorization", fmt.Sprintf("bearer %s", token))
		return c.do(req)
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

	c.sendOpts["authorization"] = func(h http.Header) {
		h.Set("authorization", fmt.Sprintf("bearer %s", t))
	}
	return t, nil
}
