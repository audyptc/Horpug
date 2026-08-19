// Package lineapi is a minimal client for the parts of LINE's platform this
// app needs: verifying a LIFF ID token (to learn a tenant's LINE userId
// during account linking) and pushing a message through the Messaging API
// (to notify a tenant once linked). See:
// https://developers.line.biz/en/reference/line-login/#verify-id-token
// https://developers.line.biz/en/reference/messaging-api/#send-push-message
package lineapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var ErrNotConfigured = errors.New("line integration is not configured")

type Client struct {
	channelAccessToken string
	channelID          string
	httpClient         *http.Client
}

func New(channelAccessToken, channelID string) *Client {
	return &Client{
		channelAccessToken: channelAccessToken,
		channelID:          channelID,
		httpClient:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) configured() bool {
	return c.channelAccessToken != "" && c.channelID != ""
}

type verifyIDTokenResponse struct {
	Sub   string `json:"sub"`
	Error string `json:"error"`
}

// VerifyIDToken confirms idToken was issued by this app's LINE channel and
// returns the LINE userId (the "sub" claim) it belongs to.
func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (string, error) {
	if !c.configured() {
		return "", ErrNotConfigured
	}

	form := url.Values{
		"id_token":  {idToken},
		"client_id": {c.channelID},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.line.me/oauth2/v2.1/verify", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed verifyIDTokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("line verify: invalid response: %w", err)
	}
	if resp.StatusCode != http.StatusOK || parsed.Sub == "" {
		if parsed.Error != "" {
			return "", fmt.Errorf("line verify failed: %s", parsed.Error)
		}
		return "", fmt.Errorf("line verify failed: status %d", resp.StatusCode)
	}

	return parsed.Sub, nil
}

type pushMessageRequest struct {
	To       string            `json:"to"`
	Messages []pushTextMessage `json:"messages"`
}

type pushTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PushMessage sends a single text message to the given LINE userId through
// the Messaging API, using the channel's access token.
func (c *Client) PushMessage(ctx context.Context, userID, text string) error {
	if !c.configured() {
		return ErrNotConfigured
	}

	payload, err := json.Marshal(pushMessageRequest{
		To:       userID,
		Messages: []pushTextMessage{{Type: "text", Text: text}},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.line.me/v2/bot/message/push", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("line push failed: status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
