package plugin

import (
	"encoding/json"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// handleAuditLogs fetches the system audit trail and formats it into a DataFrame.
func (d *Datasource) handleAuditLogs(pCtx backend.PluginContext, qm queryModel, from, to int64) backend.DataResponse {
	var response backend.DataResponse

	// 1. Fetch data from the MediaHub API
	logs, err := d.client.GetAuditLogs(qm.Limit, 0, from, to)
	if err != nil {
		response.Error = err
		return response
	}

	// 2. Initialize the DataFrame
	frame := data.NewFrame("audit_logs")

	length := len(logs)
	times := make([]time.Time, length)
	ids := make([]int64, length)
	actions := make([]string, length)
	actors := make([]string, length)
	resources := make([]string, length)
	details := make([]string, length)

	// 3. Populate the rows
	for i, log := range logs {
		// Convert Unix timestamp to UTC time.Time as expected by Grafana
		times[i] = time.UnixMilli(log.Timestamp).UTC()
		ids[i] = int64(log.ID)
		actions[i] = log.Action
		actors[i] = log.Actor
		resources[i] = log.Resource

		// Stringify the arbitrary JSON details object for clean table display
		if len(log.Details) > 0 {
			detailsBytes, _ := json.Marshal(log.Details)
			details[i] = string(detailsBytes)
		} else {
			details[i] = "{}" // Fallback for empty details
		}
	}

	// 4. Append columns to the frame
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
		data.NewField("id", nil, ids),
		data.NewField("action", nil, actions),
		data.NewField("actor", nil, actors),
		data.NewField("resource", nil, resources),
		data.NewField("details", nil, details),
	)

	response.Frames = append(response.Frames, frame)
	return response
}
