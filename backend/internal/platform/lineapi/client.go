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

// ErrUserUnreachable means the target has not added this OA as a friend, or
// has blocked it — LINE excludes such users from the friend graph, so any
// message pushed to them is silently dropped even though the Push Message
// API itself still reports success (LINE never confirms delivery, and
// specifically won't reveal block status through the push endpoint, to keep
// that private from bot developers). See:
// https://developers.line.biz/en/reference/messaging-api/#get-profile
var ErrUserUnreachable = errors.New("line user has not added the OA as a friend or has blocked it")

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

// IsFriend reports whether userID currently has this OA as a friend (i.e. a
// push message to them would actually be delivered). It calls the Get
// Profile endpoint, which LINE documents as returning 404 exactly when the
// user hasn't added the bot as a friend or has since blocked it — the only
// reliable signal LINE exposes for this, since Push Message itself always
// reports success.
func (c *Client) IsFriend(ctx context.Context, userID string) (bool, error) {
	if !c.configured() {
		return false, ErrNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.line.me/v2/bot/profile/"+url.PathEscape(userID), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("line get profile failed: status %d: %s", resp.StatusCode, string(body))
	}
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
