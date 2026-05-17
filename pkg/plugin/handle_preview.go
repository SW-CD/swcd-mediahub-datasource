package plugin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func (d *Datasource) handlePreview(pCtx backend.PluginContext, qm queryModel, from, to int64) backend.DataResponse {
	var response backend.DataResponse
	var entryID int
	var err error

	// 1. Determine the target Entry ID based on user selection
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

	// 2. Fetch the metadata to get the actual Timestamp!
	entryMeta, err := d.client.GetEntry(qm.DatabaseID, entryID)
	if err != nil {
		response.Error = fmt.Errorf("failed to fetch entry metadata: %w", err)
		return response
	}

	// 3. Fetch or Generate the Preview Data
	var previewValue string
	if qm.Base64 {
		previewJSON, err := d.client.GetEntryPreviewJSON(qm.DatabaseID, entryID)
		if err != nil {
			response.Error = fmt.Errorf("failed to fetch base64 preview: %w", err)
			return response
		}
		previewValue = previewJSON.Data
	} else {
		previewValue = fmt.Sprintf("/api/datasources/uid/%s/resources/preview/%s/%d", pCtx.DataSourceInstanceSettings.UID, qm.DatabaseID, entryID)
	}

	// 4. Construct the DataFrame
	frame := data.NewFrame("preview")

	// ADD TIME FIRST
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []time.Time{time.UnixMilli(entryMeta.Timestamp).UTC()}))
	frame.Fields = append(frame.Fields, data.NewField("id", nil, []int64{int64(entryID)}))

	previewField := data.NewField("preview", nil, []string{previewValue})
	previewField.Config = &data.FieldConfig{
		Custom: map[string]interface{}{
			"displayMode": "image",
		},
	}
	frame.Fields = append(frame.Fields, previewField)

	response.Frames = append(response.Frames, frame)
	return response
}
