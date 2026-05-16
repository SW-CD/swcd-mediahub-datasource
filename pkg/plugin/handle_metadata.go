package plugin

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// handleMetadataTable fetches entries and converts them into a standard Grafana DataFrame.
func (d *Datasource) handleMetadataTable(pCtx backend.PluginContext, qm queryModel, from, to int64) backend.DataResponse {
	var response backend.DataResponse

	// Fetch data from your MediaHub API
	entries, err := d.client.GetEntries(qm.DatabaseID, qm.Limit, from, to)
	if err != nil {
		response.Error = err
		return response
	}

	frame := data.NewFrame("metadata")

	// Initialize column slices
	length := len(entries)
	times := make([]time.Time, length)
	ids := make([]int64, length)
	filenames := make([]string, length)
	filesizes := make([]int64, length)
	mimetypes := make([]string, length)
	statuses := make([]string, length)

	// Optional Link Columns
	var entryLinks []string
	if qm.AddEntryLink {
		entryLinks = make([]string, length)
	}

	var previewLinks []string
	if qm.AddPreviewLink {
		previewLinks = make([]string, length)
	}

	// Dynamic Custom Fields: Find all unique custom keys across the dataset
	customKeys := make(map[string]bool)
	for _, e := range entries {
		for k := range e.CustomFields {
			customKeys[k] = true
		}
	}

	// Prepare slices for each custom field
	customColumns := make(map[string][]string)
	for k := range customKeys {
		customColumns[k] = make([]string, length)
	}

	// Populate the rows
	for i, e := range entries {
		times[i] = time.UnixMilli(e.Timestamp).UTC()
		ids[i] = int64(e.ID)
		filenames[i] = e.Filename
		filesizes[i] = e.Filesize
		mimetypes[i] = e.MimeType
		statuses[i] = e.Status

		if qm.AddEntryLink {
			// Construct the dynamic proxy URL for the raw file
			entryLinks[i] = fmt.Sprintf("/api/datasources/uid/%s/resources/file/%s/%d", pCtx.DataSourceInstanceSettings.UID, qm.DatabaseID, e.ID)
		}

		if qm.AddPreviewLink {
			// Construct the dynamic proxy URL for the preview image
			previewLinks[i] = fmt.Sprintf("/api/datasources/uid/%s/resources/preview/%s/%d", pCtx.DataSourceInstanceSettings.UID, qm.DatabaseID, e.ID)
		}

		// Populate custom fields (converting all values to strings for safety in tables)
		for k := range customKeys {
			if val, exists := e.CustomFields[k]; exists {
				customColumns[k][i] = fmt.Sprintf("%v", val)
			} else {
				customColumns[k][i] = "" // Empty string if the field doesn't exist on this specific entry
			}
		}
	}

	// Append core fields to the frame
	frame.Fields = append(frame.Fields,
		data.NewField("timestamp", nil, times),
		data.NewField("id", nil, ids),
		data.NewField("filename", nil, filenames),
		data.NewField("filesize", nil, filesizes),
		data.NewField("mime_type", nil, mimetypes),
		data.NewField("status", nil, statuses),
	)

	// Append custom fields
	for k, col := range customColumns {
		frame.Fields = append(frame.Fields, data.NewField("custom_"+k, nil, col))
	}

	// Append link columns with UI hints
	if qm.AddPreviewLink {
		previewField := data.NewField("preview_link", nil, previewLinks)
		previewField.Config = &data.FieldConfig{
			Custom: map[string]interface{}{
				"displayMode": "image", // Forces the Table panel to render an <img> tag
			},
		}
		frame.Fields = append(frame.Fields, previewField)
	}

	if qm.AddEntryLink {
		entryField := data.NewField("entry_link", nil, entryLinks)
		entryField.Config = &data.FieldConfig{
			// Automatically turns the cell text into a clickable link
			Links: []data.DataLink{
				{
					Title:       "Open File",
					URL:         "${__value.raw}", // Grafana variable pointing to the cell's text
					TargetBlank: true,             // Opens in a new tab
				},
			},
		}
		frame.Fields = append(frame.Fields, entryField)
	}

	response.Frames = append(response.Frames, frame)
	return response
}
