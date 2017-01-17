package apns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/http2"
)

// APN service endpoint URLs.
const (
	DevelopmentGateway = "https://api.development.push.apple.com"
	ProductionGateway  = "https://api.push.apple.com"
)

// Client represents the Apple Push Notification Service that you send notifications to.
type Client struct {
	http     *http.Client
	sendOpts []SendOption
	endpoint string
}

// NewClient creates new AONS client based on defined Options.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		http: &http.Client{
			Transport: &http.Transport{},
		},
		endpoint: ProductionGateway,
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

// SendWithRetry sends Notification to the APN service and retries in case
// network or server error.
func (c *Client) SendWithRetry(token string, p Payload, maxAttempts int, opts ...SendOption) (*Response, error) {
	req, err := c.prepareRequest(token, p, opts...)
	if err != nil {
		return nil, err
	}
	resp := new(Response)
	err = retry(func() error {
		var err error
		resp, err = c.do(req)
		return err
	}, maxAttempts)
	if err != nil {
		return nil, err
	}

	return resp, nil
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

	response := new(Response)
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}
	response.NotificationID = resp.Header.Get("apns-id")
	return response, response.Error
}
