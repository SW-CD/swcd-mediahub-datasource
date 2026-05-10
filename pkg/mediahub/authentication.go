package mediahub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// doAuthenticatedRequest handles attaching the Bearer token and executing the request.
func (c *mediahubClient) doAuthenticatedRequest(method, endpoint string, payload []byte) ([]byte, error) {
	token, err := c.getValidToken()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewBuffer(payload)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// getValidToken ensures an active access token is available, refreshing or logging in if necessary.
func (c *mediahubClient) getValidToken() (string, error) {
	// 1. Fast path: check if we have a valid token (using a read lock)
	c.tokenMutex.RLock()
	// Add a 10-second buffer to the expiry to prevent race conditions during the HTTP call
	if c.accessToken != "" && time.Now().Add(10*time.Second).Before(c.tokenExpiry) {
		token := c.accessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	// 2. Slow path: acquire write lock to perform token fetching
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Double-check pattern: another goroutine might have refreshed the token while we waited for the lock
	if c.accessToken != "" && time.Now().Add(10*time.Second).Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	// 3. Attempt Refresh if we have a refresh token
	if c.refreshToken != "" {
		err := c.performTokenRefresh()
		if err == nil {
			return c.accessToken, nil
		}
		// If refresh fails (e.g., revoked or expired), fall through to full login
	}

	// 4. Perform full Basic Auth login
	err := c.performLogin()
	if err != nil {
		return "", err
	}

	return c.accessToken, nil
}

// performLogin exchanges username/password for a new token pair via Basic Auth.
func (c *mediahubClient) performLogin() error {
	req, err := http.NewRequest("POST", c.baseURL+"/api/token", nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	return c.executeTokenRequest(req)
}

// performTokenRefresh exchanges an existing refresh token for a new token pair.
func (c *mediahubClient) performTokenRefresh() error {
	refreshReq := TokenRefreshRequest{
		RefreshToken: c.refreshToken,
	}

	payload, err := json.Marshal(refreshReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/token/refresh", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return c.executeTokenRequest(req)
}

// executeTokenRequest is a shared helper for parsing the token response.
func (c *mediahubClient) executeTokenRequest(req *http.Request) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.refreshToken = tokenResp.RefreshToken

	// Determine the exact expiration from the token itself
	expTime, err := parseTokenExpiry(c.accessToken)
	if err != nil {
		// Fallback to a safe default if the token is malformed
		c.tokenExpiry = time.Now().Add(4*time.Minute + 30*time.Second)
	} else {
		// Subtract a 30-second buffer to prevent race conditions during network latency
		c.tokenExpiry = expTime.Add(-30 * time.Second)
	}

	return nil
}

// parseTokenExpiry decodes the JWT payload to extract the 'exp' claim.
func parseTokenExpiry(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid JWT format")
	}

	// JWTs use Base64Url encoding without padding
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return time.Time{}, fmt.Errorf("failed to unmarshal JWT claims: %w", err)
	}

	if claims.Exp == 0 {
		return time.Time{}, fmt.Errorf("JWT does not contain an exp claim")
	}

	// Convert the Unix epoch timestamp (seconds) into a time.Time object
	return time.Unix(claims.Exp, 0), nil
}
