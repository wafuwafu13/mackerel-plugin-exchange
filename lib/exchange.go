package exchange

import (
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

// ExchangePlugin mackerel plugin
type ExchangePlugin struct {
	Prefix string
}

// MetricKeyPrefix interface for PluginWithPrefix
func (u ExchangePlugin) MetricKeyPrefix() string {
	if u.Prefix == "" {
		u.Prefix = "exchange"
	}
	return u.Prefix
}

// GraphDefinition interface for mackerelplugin
func (u ExchangePlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(u.Prefix)
	return map[string]mp.Graphs{
		"": {
			Label: labelPrefix,
			Unit:  mp.UnitFloat,
			Metrics: []mp.Metrics{
				{Name: "USD", Label: "USD"},
				{Name: "EUR", Label: "EUR"},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (u ExchangePlugin) FetchMetrics() (map[string]float64, error) {
	return map[string]float64{"USD": 145.1, "EUR": 158.6}, nil
}

// Do the plugin
func Do() {
	u := ExchangePlugin{}
	helper := mp.NewMackerelPlugin(u)
	helper.Run()
}
