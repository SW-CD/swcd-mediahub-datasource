package plugin

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// handleConfigMap provides the React frontend with the user's admin status and available databases.
// The frontend calls this once when the Query Editor mounts.
func (d *Datasource) handleConfigMap(w http.ResponseWriter, r *http.Request) {
	user, err := d.client.GetMe()
	if err != nil {
		http.Error(w, "failed to get user profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dbs, err := d.client.GetDatabases()
	if err != nil {
		http.Error(w, "failed to fetch databases: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"is_admin":  user.IsAdmin,
		"databases": dbs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleFileProxy intercepts the media request, enforces the size limit, and proxies the binary stream.
// Expected URL path: /file/{databaseID}/{entryID}?max_size=4
func (d *Datasource) handleFileProxy(w http.ResponseWriter, r *http.Request) {
	// 1. Parse the URL path to extract IDs
	path := strings.TrimPrefix(r.URL.Path, "/file/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.Error(w, "invalid path format, expected /file/{databaseID}/{entryID}", http.StatusBadRequest)
		return
	}

	databaseID := parts[0]
	entryID, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "invalid entry ID", http.StatusBadRequest)
		return
	}

	// 2. Check the max_size query parameter (default 4MB)
	maxSizeMB := 4.0
	if sizeStr := r.URL.Query().Get("max_size"); sizeStr != "" {
		if parsedSize, err := strconv.ParseFloat(sizeStr, 64); err == nil {
			maxSizeMB = parsedSize
		}
	}
	maxSizeBytes := int64(maxSizeMB * 1024 * 1024)

	// 3. Enforce the size limit by fetching metadata first
	meta, err := d.client.GetEntryMetadata(databaseID, entryID)
	if err != nil {
		http.Error(w, "failed to fetch metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if maxSizeBytes > 0 && meta.Filesize > maxSizeBytes {
		http.Error(w, "file exceeds the configured maximum size limit", http.StatusRequestEntityTooLarge)
		return
	}

	// 4. Proxy the file stream
	resp, err := d.client.ProxyEntryFile(databaseID, entryID, r.Header)
	if err != nil {
		http.Error(w, "failed to proxy file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Clear any default headers Grafana might have set (like an incorrect Content-Type)
	for k := range w.Header() {
		w.Header().Del(k)
	}

	// Copy all headers from the MediaHub response to Grafana
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	// Stream the body chunk by chunk
	io.Copy(w, resp.Body)
}

// handlePreviewProxy intercepts the preview request and streams the raw binary image.
// Expected URL path: /preview/{databaseID}/{entryID}
func (d *Datasource) handlePreviewProxy(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/preview/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.Error(w, "invalid path format, expected /preview/{databaseID}/{entryID}", http.StatusBadRequest)
		return
	}

	databaseID := parts[0]
	entryID, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "invalid entry ID", http.StatusBadRequest)
		return
	}

	// Stream the preview directly as binary
	resp, err := d.client.ProxyEntryPreview(databaseID, entryID, r.Header)
	if err != nil {
		http.Error(w, "failed to proxy preview: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Clear any default headers Grafana might have set
	for k := range w.Header() {
		w.Header().Del(k)
	}

	// Copy all headers (like Content-Type: image/webp and Cache-Control)
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// handleVariableEntries fetches a lightweight list of entries specifically for the Grafana variables dropdown.
func (d *Datasource) handleVariableEntries(w http.ResponseWriter, r *http.Request) {
	dbID := strings.TrimPrefix(r.URL.Path, "/variables/entries/")
	if dbID == "" {
		http.Error(w, "missing database ID", http.StatusBadRequest)
		return
	}

	// Fetch the latest 1000 entries to populate the dropdown (bypassing time filters via 0, 0)
	entries, err := d.client.GetEntries(dbID, 1000, 0, 0)
	if err != nil {
		http.Error(w, "failed to fetch entries: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// We only need the ID and Filename to populate the Grafana dropdown, so we construct a lightweight struct
	type VariableEntry struct {
		ID       int    `json:"id"`
		Filename string `json:"filename"`
	}

	var res []VariableEntry
	for _, e := range entries {
		res = append(res, VariableEntry{
			ID:       e.ID,
			Filename: e.Filename,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
