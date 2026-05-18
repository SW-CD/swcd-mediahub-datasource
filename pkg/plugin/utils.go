package plugin

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/swcd/mediahub/pkg/mediahub"
)

// buildMetadataFields takes a slice of MediaHub entries and dynamically constructs Grafana DataFrame columns
// for all core fields, custom fields, and media fields.
func buildMetadataFields(entries []mediahub.Entry) []*data.Field {
	length := len(entries)
	times := make([]time.Time, length)
	ids := make([]int64, length)
	filenames := make([]string, length)
	filesizes := make([]int64, length)
	mimetypes := make([]string, length)
	statuses := make([]string, length)

	// 1. Discover all unique dynamic keys across the entries
	customKeys := make(map[string]bool)
	mediaKeys := make(map[string]bool)

	for _, e := range entries {
		for k := range e.CustomFields {
			customKeys[k] = true
		}
		for k := range e.MediaFields {
			mediaKeys[k] = true
		}
	}

	// 2. Initialize the dynamic columns
	customColumns := make(map[string][]string)
	for k := range customKeys {
		customColumns[k] = make([]string, length)
	}

	mediaColumns := make(map[string][]string)
	for k := range mediaKeys {
		mediaColumns[k] = make([]string, length)
	}

	// 3. Populate all rows
	for i, e := range entries {
		times[i] = time.UnixMilli(e.Timestamp).UTC()
		ids[i] = int64(e.ID)
		filenames[i] = e.Filename
		filesizes[i] = e.Filesize
		mimetypes[i] = e.MimeType
		statuses[i] = e.Status

		// Populate custom fields
		for k := range customKeys {
			if val, exists := e.CustomFields[k]; exists {
				customColumns[k][i] = fmt.Sprintf("%v", val)
			} else {
				customColumns[k][i] = ""
			}
		}

		// Populate media fields
		for k := range mediaKeys {
			if val, exists := e.MediaFields[k]; exists {
				mediaColumns[k][i] = fmt.Sprintf("%v", val)
			} else {
				mediaColumns[k][i] = ""
			}
		}
	}

	// 4. Construct the slice of data.Field
	var fields []*data.Field

	// Note: Naming the timestamp column "time" is standard Grafana practice
	// for allowing Time Series and related panels to automatically detect it.
	fields = append(fields,
		data.NewField("time", nil, times),
		data.NewField("id", nil, ids),
		data.NewField("filename", nil, filenames),
		data.NewField("filesize", nil, filesizes),
		data.NewField("mime_type", nil, mimetypes),
		data.NewField("status", nil, statuses),
	)

	// Append Custom Fields
	for k, col := range customColumns {
		fields = append(fields, data.NewField("custom_"+k, nil, col))
	}

	// Append Media Fields
	for k, col := range mediaColumns {
		fields = append(fields, data.NewField("media_"+k, nil, col))
	}

	return fields
}
