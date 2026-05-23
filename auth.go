package paypal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	accessTokenEndpoint = "/v1/oauth2/token"
)

type AccessToken struct {
	Scope     string `json:"scope"`
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	AppID     string `json:"app_id"`
	ExpiresIn int    `json:"expires_in"`
	Nonce     string `json:"nonce"`
	CreatedAt time.Time
}

func newAccessToken(ctx context.Context, hostURL, clientID, clientSecret string) (*AccessToken, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		hostURL+accessTokenEndpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("paypal auth failed: status %d: %s", res.StatusCode, string(body))
	}

	at := AccessToken{}
	if err := json.NewDecoder(res.Body).Decode(&at); err != nil {
		return nil, err
	}

	at.CreatedAt = time.Now()

	return &at, nil
}

func (a *AccessToken) expired() bool {
	if a.CreatedAt.IsZero() || a.ExpiresIn <= 0 {
		return true
	}

	// Buffer as to not used token that expires too soon.
	const expiryBuffer = 30 * time.Second

	ttl := time.Duration(a.ExpiresIn)*time.Second - expiryBuffer

	return time.Now().After(a.CreatedAt.Add(ttl))
}
