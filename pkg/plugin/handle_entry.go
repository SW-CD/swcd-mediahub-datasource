package plugin

import (
	"fmt"
	"strconv"

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
