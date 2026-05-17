package plugin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/swcd/mediahub/pkg/mediahub"
	"github.com/swcd/mediahub/pkg/models"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil) // Add this interface
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

type Datasource struct {
	client          mediahub.Client
	resourceHandler backend.CallResourceHandler // Holds the HTTP multiplexer
}

func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}

	client := mediahub.NewClient(config.URL, config.Username, config.Secrets.Password)

	// Set up the internal router for CallResource endpoints
	mux := http.NewServeMux()
	ds := &Datasource{
		client:          client,
		resourceHandler: httpadapter.New(mux),
	}

	// Register the custom routes
	mux.HandleFunc("/config", ds.handleConfigMap)
	mux.HandleFunc("/file/", ds.handleFileProxy)
	mux.HandleFunc("/preview/", ds.handlePreviewProxy)
	mux.HandleFunc("/variables/entries/", ds.handleVariableEntries) // <-- Add this

	return ds, nil
}

func (d *Datasource) Dispose() {
	// Clean up resources if needed
}

// CallResource intercepts HTTP requests from the frontend and passes them to our ServeMux router.
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return d.resourceHandler.CallResource(ctx, req, sender)
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// queryModel represents the JSON query sent from the React frontend.
type queryModel struct {
	DatabaseID     string `json:"databaseId"`
	Model          string `json:"model"`
	Limit          int    `json:"limit"`
	TStart         int64  `json:"tstart"`
	TEnd           int64  `json:"tend"`
	AddPreviewLink bool   `json:"addPreviewLink"`
	AddEntryLink   bool   `json:"addEntryLink"`

	// Fields for Previews and Entries
	TargetSelection string  `json:"targetSelection"`
	EntryID         string  `json:"entryId"`
	Base64          bool    `json:"base64"`
	MaxFileSize     float64 `json:"maxFileSize"`
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {

	// 1. Unmarshal the React query into our Go struct
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "json unmarshal: "+err.Error())
	}

	// 2. Determine time range (use frontend overrides if provided, otherwise default to dashboard range)
	from := query.TimeRange.From.UnixMilli()
	to := query.TimeRange.To.UnixMilli()
	if qm.TStart > 0 {
		from = qm.TStart
	}
	if qm.TEnd > 0 {
		to = qm.TEnd
	}

	// 3. Route the query based on the selected Model
	switch qm.Model {
	case "get metadata table":
		return d.handleMetadataTable(pCtx, qm, from, to)
	case "get preview":
		return d.handlePreview(pCtx, qm, from, to)
	case "get entry":
		return d.handleEntry(pCtx, qm, from, to)
	case "get audit logs":
		// Placeholder for next step
		return backend.ErrDataResponse(backend.StatusNotImplemented, "get audit logs not yet implemented")
	default:
		return backend.ErrDataResponse(backend.StatusNotImplemented, "model not yet implemented in backend")
	}
}

// CheckHealth handles health checks sent from Grafana to the plugin.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)

	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	// 1. Validate that all required fields are present
	if config.URL == "" {
		res.Status = backend.HealthStatusError
		res.Message = "URL is missing"
		return res, nil
	}
	if config.Username == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Username is missing"
		return res, nil
	}
	if config.Secrets.Password == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Password is missing"
		return res, nil
	}

	// 2. Actively test the connection to MediaHub using our client
	_, err = d.client.GetMe()
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Authentication failed: " + err.Error()
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working and authenticated successfully.",
	}, nil
}
