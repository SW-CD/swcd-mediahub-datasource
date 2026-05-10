package mediahub

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// GetAuditLogs fetches a list of system audit logs. Used for the "get audit logs" model.
// It requires the authenticated user to have the IsAdmin global role.
func (c *mediahubClient) GetAuditLogs(limit, offset int, tstart, tend int64) ([]AuditLog, error) {
	// Build query parameters
	q := url.Values{}

	if limit > 0 {
		q.Add("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Add("offset", strconv.Itoa(offset))
	}

	// Default to descending so the most recent events are always shown first
	q.Add("order", "desc")

	if tstart > 0 {
		q.Add("tstart", strconv.FormatInt(tstart, 10))
	}
	if tend > 0 {
		q.Add("tend", strconv.FormatInt(tend, 10))
	}

	endpoint := "/api/audit"
	if len(q) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, q.Encode())
	}

	body, err := c.doAuthenticatedRequest("GET", endpoint, nil)
	if err != nil {
		// Gracefully catch the 403 Forbidden error for non-admins.
		// We intercept the generic status error from doAuthenticatedRequest
		// and return the user-friendly message specified in the concept.
		if strings.Contains(err.Error(), "status 403") {
			return nil, fmt.Errorf("Error: Admin privileges required to view audit logs")
		}
		return nil, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	var logs []AuditLog
	if err := json.Unmarshal(body, &logs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit logs: %w", err)
	}

	return logs, nil
}
