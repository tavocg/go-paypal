// Package paypal provides a minimal client for the PayPal REST API.
package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"unsafe"
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
		c.clientSecret = clientSecret
	}
}

func NewClient(ctx context.Context, opts ...ClientOpt) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	if err := clientFieldsFromEnv(c); err != nil {
		return nil, err
	}

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

func clientFieldsFromEnv(c *Client) error {
	if c == nil {
		return fmt.Errorf("client must be non-nil")
	}

	elem := reflect.ValueOf(c).Elem()

	elemType := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		fieldType := elemType.Field(i)
		envVar := fieldType.Tag.Get("env")
		if envVar == "" {
			continue
		}

		field := elem.Field(i)
		if field.Kind() != reflect.String {
			return fmt.Errorf("env tag only supported on string fields: %s", fieldType.Name)
		}
		currentValue := field.String()
		if currentValue != "" {
			continue
		}

		if envValue, ok := os.LookupEnv(envVar); ok {
			if field.CanSet() {
				field.SetString(envValue)
				continue
			}
			if !field.CanAddr() {
				return fmt.Errorf("cannot address field %s", fieldType.Name)
			}

			reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().SetString(envValue)
		}
	}

	return nil
}

func (c *Client) request(ctx context.Context, method, endpoint string, payload any, dest any) error {
	if err := c.checkAT(ctx); err != nil {
		return err
	}

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.hostURL+endpoint, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.currentAT.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("paypal request failed: status %d: %s", res.StatusCode, string(body))
	}

	if dest == nil {
		return nil
	}

	return json.NewDecoder(res.Body).Decode(dest)
}
