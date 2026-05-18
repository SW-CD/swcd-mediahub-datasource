package plugin

import (
	"fmt"

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

	// 1. Generate all standard metadata fields using our shared utility
	frame.Fields = append(frame.Fields, buildMetadataFields(entries)...)

	// 2. Generate Link Columns (if requested by the user)
	length := len(entries)
	if qm.AddPreviewLink || qm.AddEntryLink {
		var entryLinks, previewLinks []string

		if qm.AddEntryLink {
			entryLinks = make([]string, length)
		}
		if qm.AddPreviewLink {
			previewLinks = make([]string, length)
		}

		for i, e := range entries {
			if qm.AddEntryLink {
				entryLinks[i] = fmt.Sprintf("/api/datasources/uid/%s/resources/file/%s/%d?max_size=%f",
					pCtx.DataSourceInstanceSettings.UID,
					qm.DatabaseID,
					e.ID,
					qm.MaxFileSize,
				)
			}
			if qm.AddPreviewLink {
				previewLinks[i] = fmt.Sprintf("/api/datasources/uid/%s/resources/preview/%s/%d",
					pCtx.DataSourceInstanceSettings.UID,
					qm.DatabaseID,
					e.ID,
				)
			}
		}

		// Append link columns with UI hints
		if qm.AddPreviewLink {
			previewField := data.NewField("preview_link", nil, previewLinks)
			previewField.Config = &data.FieldConfig{
				Custom: map[string]interface{}{
					"displayMode": "image",
				},
			}
			frame.Fields = append(frame.Fields, previewField)
		}

		if qm.AddEntryLink {
			entryField := data.NewField("entry_link", nil, entryLinks)
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
		}
	}

	response.Frames = append(response.Frames, frame)
	return response
}
