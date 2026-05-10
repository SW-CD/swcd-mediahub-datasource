package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// PluginSettings holds the plaintext configuration.
type PluginSettings struct {
	URL      string                `json:"url"`
	Username string                `json:"username"`
	Secrets  *SecretPluginSettings `json:"-"`
}

// SecretPluginSettings holds the sensitive configuration stored in Grafana's secure credential manager.
type SecretPluginSettings struct {
	Password string `json:"password"`
}

// LoadPluginSettings deserializes the Grafana instance settings.
func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)

	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		Password: source["password"], // Matches the frontend SecureJSONData key we will define later
	}
}
