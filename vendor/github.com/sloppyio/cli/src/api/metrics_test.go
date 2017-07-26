package api

import (
	"reflect"
	"testing"
	"time"
)

func TestMetricsUnmarshal(t *testing.T) {
	input := []byte(`[{"metric":"container_memory_usage_bytes","values":[{"name":"node-1234", "data": [{"x":1446646639934,"y":31244288}]}]}]`)
	want := Metrics{
		"container_memory_usage_bytes": Series{
			"node-1234": DataPoints{
				&Point{
					X: Timestamp{time.Date(2015, 11, 4, 14, 17, 19, 0, time.UTC)},
					Y: Float64(31244288),
				},
			},
		},
	}

	metrics := make(Metrics)
	if err := metrics.UnmarshalJSON(input); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(metrics, want) {
		t.Errorf("Unmarshal(%v) = %v, %v", string(input), metrics, want)
	}
}

func TestMetricsUnmarshal_invalidJSONBody(t *testing.T) {
	input := []byte(`{`)

	metrics := make(Metrics)
	if err := metrics.UnmarshalJSON(input); err == nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
