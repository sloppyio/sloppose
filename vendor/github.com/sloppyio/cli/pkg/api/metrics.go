package api

import (
	"encoding/json"
	"fmt"
)

// Metrics represents all sloppy stats.
type Metrics map[string]Series

// Series represents a named set of datapoints.
type Series map[string]DataPoints

// DataPoints represents all values of a serie.
type DataPoints []*DataPoint

// DataPoint represents a value at specific time
type DataPoint struct {
	X Timestamp `json:"x,omitempty"`
	Y *float64  `json:"y,omitempty"`
}

// UnmarshalJSON decodes sloppy's metric format.
func (m Metrics) UnmarshalJSON(data []byte) error {
	var aux = []struct {
		Name   string `json:"metric,omitempty"`
		Series []struct {
			Name   string     `json:"name,omitempty"`
			Values DataPoints `json:"data,omitempty"`
		} `json:"values,omitempty"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("metrics couldn't decode, got %v", err)
	}

	for _, metric := range aux {
		series := make(Series, len(metric.Series))
		for _, source := range metric.Series {
			series[source.Name] = source.Values
		}
		m[metric.Name] = series
	}

	return nil
}
