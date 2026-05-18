package mediahub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// GetEntries fetches a list of entry metadata. Used for the "get metadata" model.
func (c *mediahubClient) GetEntries(databaseID string, limit int, tstart, tend int64) ([]Entry, error) {
	// Build query parameters
	q := url.Values{}
	q.Add("limit", strconv.Itoa(limit))
	q.Add("order", "desc")
	if tstart > 0 {
		q.Add("tstart", strconv.FormatInt(tstart, 10))
	}
	if tend > 0 {
		q.Add("tend", strconv.FormatInt(tend, 10))
	}

	endpoint := fmt.Sprintf("/api/database/%s/entries?%s", databaseID, q.Encode())

	body, err := c.doAuthenticatedRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entries: %w", err)
	}

	var entries []Entry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entries: %w", err)
	}

	return entries, nil
}

// GetEntry retrieves all metadata for a single entry.
func (c *mediahubClient) GetEntry(databaseID string, entryID int) (*Entry, error) {
	endpoint := fmt.Sprintf("/api/database/%s/entry/%d", databaseID, entryID)

	// Use your established authentication handler
	respBytes, err := c.doAuthenticatedRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var entry Entry
	if err := json.Unmarshal(respBytes, &entry); err != nil {
		return nil, fmt.Errorf("failed to decode entry JSON: %w", err)
	}

	return &entry, nil
}

// GetEntryFileJSON retrieves the file as a Base64-encoded string wrapped in JSON.
func (c *mediahubClient) GetEntryFileJSON(databaseID string, entryID int) (*FileJSONResponse, error) {
	endpoint := fmt.Sprintf("/api/database/%s/entry/%d/file", databaseID, entryID)

	// We use getValidToken directly instead of doAuthenticatedRequest
	// because we MUST inject the "Accept: application/json" header
	// to trigger the API's Content Negotiation.
	token, err := c.getValidToken()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json") // Triggers the Base64 JSON response

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var fileResp FileJSONResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		return nil, fmt.Errorf("failed to decode base64 file JSON: %w", err)
	}

	return &fileResp, nil
}

// GetLatestEntryID handles step one of the "get last" and "get last in range" options.
func (c *mediahubClient) GetLatestEntryID(databaseID string, tstart, tend int64) (int, error) {
	entries, err := c.GetEntries(databaseID, 1, tstart, tend)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, fmt.Errorf("no entries found in the specified time range")
	}
	return entries[0].ID, nil
}

// GetEntryMetadata retrieves metadata for a single specific entry.
func (c *mediahubClient) GetEntryMetadata(databaseID string, entryID int) (*Entry, error) {
	endpoint := fmt.Sprintf("/api/database/%s/entry/%d", databaseID, entryID)

	body, err := c.doAuthenticatedRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entry metadata: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry metadata: %w", err)
	}

	return &entry, nil
}

// GetEntryPreviewJSON fetches the Base64 encoded preview image for the "get preview" model.
func (c *mediahubClient) GetEntryPreviewJSON(databaseID string, entryID int) (*PreviewResponse, error) {
	endpoint := fmt.Sprintf("/api/database/%s/entry/%d/preview", databaseID, entryID)

	// Since doAuthenticatedRequest doesn't currently allow custom headers,
	// we handle the custom Accept header directly to trigger the Base64 JSON response.
	token, err := c.getValidToken()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json") // CRITICAL: Triggers Base64 instead of binary

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var preview PreviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		return nil, fmt.Errorf("failed to decode preview JSON: %w", err)
	}

	return &preview, nil
}

// ProxyEntryFile stream-proxies the raw binary file for the "get entry" CallResource endpoint.
func (c *mediahubClient) ProxyEntryFile(databaseID string, entryID int, incomingHeaders http.Header) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/database/%s/entry/%d/file", databaseID, entryID)

	token, err := c.getValidToken()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Only extract and forward the 'Range' header.
	if rangeHeader := incomingHeaders.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	// Override with our internal auth and ensure standard binary response
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "*/*")

	return c.httpClient.Do(req)
}

// ProxyEntryPreview stream-proxies the raw binary preview image.
func (c *mediahubClient) ProxyEntryPreview(databaseID string, entryID int, incomingHeaders http.Header) (*http.Response, error) {
	endpoint := fmt.Sprintf("/api/database/%s/entry/%d/preview", databaseID, entryID)

	token, err := c.getValidToken()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// For previews, we generally don't need any frontend headers except maybe caching.
	// We specifically avoid copying the whole map to prevent side-effects.
	if ifNoneMatch := incomingHeaders.Get("If-None-Match"); ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "*/*")

	return c.httpClient.Do(req)
}
