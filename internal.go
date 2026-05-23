package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
