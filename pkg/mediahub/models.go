package mediahub

// Auth Models
type TokenRequest struct {
	IDPToken string `json:"idp_token,omitempty"`
}

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type User struct {
	ID          int          `json:"id"`
	Username    string       `json:"username"`
	IsAdmin     bool         `json:"is_admin"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	DatabaseID string `json:"database_id"`
	CanView    bool   `json:"can_view"`
	CanCreate  bool   `json:"can_create"`
	CanEdit    bool   `json:"can_edit"`
	CanDelete  bool   `json:"can_delete"`
}

// Database Models
type Database struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	ContentType string        `json:"content_type"`
	Stats       DatabaseStats `json:"stats"`
	// Add other fields (custom_fields, config) as needed later
}

type DatabaseStats struct {
	EntryCount          int   `json:"entry_count"`
	TotalDiskSpaceBytes int64 `json:"total_disk_space_bytes"`
}

// Entry Models
type Entry struct {
	DatabaseID   string                 `json:"database_id"`
	Timestamp    int64                  `json:"timestamp"`
	ID           int                    `json:"id"`
	Filesize     int64                  `json:"filesize"`
	MimeType     string                 `json:"mime_type"`
	Filename     string                 `json:"filename"`
	Status       string                 `json:"status"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type PreviewResponse struct {
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"` // The base64 string
}

type FileJSONResponse struct {
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Data     string `json:"data"`
}

// AuditLog represents a single system audit event.
type AuditLog struct {
	ID        int                    `json:"id"`
	Timestamp int64                  `json:"timestamp"`
	Action    string                 `json:"action"`
	Actor     string                 `json:"actor"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details,omitempty"`
}
