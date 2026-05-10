package mediahub

import (
	"encoding/json"
	"fmt"
)

// GetDatabases retrieves the list of databases the user has access to.
func (c *mediahubClient) GetDatabases() ([]Database, error) {
	body, err := c.doAuthenticatedRequest("GET", "/api/databases", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch databases: %w", err)
	}

	var databases []Database
	if err := json.Unmarshal(body, &databases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal database array: %w", err)
	}

	return databases, nil
}
