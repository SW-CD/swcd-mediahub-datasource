package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/swcd/mediahub/pkg/mediahub"
)

func (d *Datasource) handleEntry(pCtx backend.PluginContext, qm queryModel, from, to int64) backend.DataResponse {
	var response backend.DataResponse
	var entryID int
	var err error

	// 1. Determine target ID
	switch qm.TargetSelection {
	case "get ID":
		if qm.EntryID == "" {
			response.Error = fmt.Errorf("entry ID cannot be empty")
			return response
		}
		entryID, err = strconv.Atoi(qm.EntryID)
		if err != nil {
			response.Error = fmt.Errorf("invalid entry ID format: %v", err)
			return response
		}
	case "get last":
		entryID, err = d.client.GetLatestEntryID(qm.DatabaseID, 0, 0)
		if err != nil {
			response.Error = fmt.Errorf("failed to get latest entry: %w", err)
			return response
		}
	case "get last in range":
		entryID, err = d.client.GetLatestEntryID(qm.DatabaseID, from, to)
		if err != nil {
			response.Error = fmt.Errorf("failed to get latest entry in range: %w", err)
			return response
		}
	default:
		response.Error = fmt.Errorf("unknown target selection: %s", qm.TargetSelection)
		return response
	}

	// 2. Fetch metadata for Timestamp, Size Checking, and general fields
	entryMeta, err := d.client.GetEntry(qm.DatabaseID, entryID)
	if err != nil {
		response.Error = fmt.Errorf("failed to fetch entry metadata: %w", err)
		return response
	}

	// 3. SAFETY CHECK: Prevent massive files from crashing the frontend
	if qm.MaxFileSize > 0 {
		maxBytes := int64(qm.MaxFileSize * 1024 * 1024)
		if entryMeta.Filesize > maxBytes {
			response.Error = fmt.Errorf("file size (%.2f MB) exceeds the configured maximum limit (%.2f MB)", float64(entryMeta.Filesize)/(1024*1024), qm.MaxFileSize)
			return response
		}
	}

	// 4. Generate the Entry Data
	var entryValue string
	if qm.Base64 {
		fileJSON, err := d.client.GetEntryFileJSON(qm.DatabaseID, entryID)
		if err != nil {
			response.Error = fmt.Errorf("failed to fetch base64 file: %w", err)
			return response
		}
		entryValue = fileJSON.Data
	} else {
		entryValue = fmt.Sprintf("/api/datasources/uid/%s/resources/file/%s/%d?max_size=%f",
			pCtx.DataSourceInstanceSettings.UID,
			qm.DatabaseID,
			entryID,
			qm.MaxFileSize,
		)
	}

	// 5. Construct the DataFrame
	frame := data.NewFrame("entry")

	// Append ALL metadata fields using our shared utility function
	frame.Fields = append(frame.Fields, buildMetadataFields([]mediahub.Entry{*entryMeta})...)

	// Append the dedicated Entry field (URL or Base64) with clickable links
	entryField := data.NewField("entry", nil, []string{entryValue})
	entryField.Config = &data.FieldConfig{
		Links: []data.DataLink{
			{
				Title:       "Open File",
				URL:         "${__value.raw}",
				TargetBlank: true,
			},
		},
	}
	frame.Fields = append(frame.Fields, entryField)

	response.Frames = append(response.Frames, frame)
	return response
}

// handleVariableEntries fetches a lightweight list of entries specifically for the Grafana variables dropdown.
func (d *Datasource) handleVariableEntries(w http.ResponseWriter, r *http.Request) {
	dbID := strings.TrimPrefix(r.URL.Path, "/variables/entries/")
	if dbID == "" {
		http.Error(w, "missing database ID", http.StatusBadRequest)
		return
	}

	// 1. Read optional time boundaries from the URL query params
	var from, to int64
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if val, err := strconv.ParseInt(fromStr, 10, 64); err == nil {
			from = val
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if val, err := strconv.ParseInt(toStr, 10, 64); err == nil {
			to = val
		}
	}

	// 2. Pass the parsed time parameters to GetEntries (instead of hardcoding 0, 0)
	entries, err := d.client.GetEntries(dbID, 1000, from, to)
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
