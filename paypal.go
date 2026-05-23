// Package paypal <description here>
package paypal

import (
	"context"
	"fmt"
)

const (
	SandboxHost    = "https://api-m.sandbox.paypal.com"
	ProductionHost = "https://api-m.paypal.com"
)

type Client struct {
	hostURL      string
	clientID     string `env:"PAYPAL_CLIENT_ID"`
	clientSecret string `env:"PAYPAL_CLIENT_SECRET"`
	currentAT    *AccessToken
}

type ClientOpt func(*Client)

func WithHost(url string) ClientOpt {
	return func(c *Client) {
		c.hostURL = url
	}
}

func WithClientID(clientID string) ClientOpt {
	return func(c *Client) {
		c.clientID = clientID
	}
}

func WithClientSecret(clientSecret string) ClientOpt {
	return func(c *Client) {
		c.clientID = clientSecret
	}
}

func NewClient(ctx context.Context, opts ...ClientOpt) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	// TODO: (create and) run function clientFieldsFromEnv(c *Client) that
	// uses reflect to set the field from the env:"ENV_VAR" tag of the struct

	if c.hostURL == "" || c.clientID == "" || c.clientSecret == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	return c, nil
}

// checkAT refreshes the client's access token if needed, errors on failure.
func (c *Client) checkAT(ctx context.Context) error {
	if c.currentAT == nil || c.currentAT.expired() {
		at, err := newAccessToken(ctx, c.hostURL, c.clientID, c.clientSecret)
		if err != nil {
			return err
		}
		c.currentAT = at
	}
	return nil
}
