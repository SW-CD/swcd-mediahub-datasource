package mediahub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Client interface defines the available operations against the Mediahub API.
type Client interface {
	GetMe() (*User, error)
	GetDatabases() ([]Database, error)

	// Entries
	GetEntry(databaseID string, entryID int) (*Entry, error)
	GetEntries(databaseID string, limit int, tstart, tend int64) ([]Entry, error)
	GetLatestEntryID(databaseID string, tstart, tend int64) (int, error)
	GetEntryMetadata(databaseID string, entryID int) (*Entry, error)
	GetEntryFileJSON(databaseID string, entryID int) (*FileJSONResponse, error)
	GetEntryPreviewJSON(databaseID string, entryID int) (*PreviewResponse, error)
	ProxyEntryPreview(databaseID string, entryID int, incomingHeaders http.Header) (*http.Response, error) // <-- Add this
	ProxyEntryFile(databaseID string, entryID int, incomingHeaders http.Header) (*http.Response, error)

	// Audit Logs
	GetAuditLogs(limit, offset int, tstart, tend int64) ([]AuditLog, error)
}

// mediahubClient implements the Client interface.
type mediahubClient struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string

	// Token management
	tokenMutex   sync.RWMutex
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
}

// NewClient creates a new API client.
func NewClient(baseURL, username, password string) Client {
	return &mediahubClient{
		httpClient: &http.Client{
			Timeout: time.Second * 30, // Sensible default for API calls
		},
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

// GetMe retrieves the current user's profile and database permissions.
func (c *mediahubClient) GetMe() (*User, error) {
	body, err := c.doAuthenticatedRequest("GET", "/api/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}
