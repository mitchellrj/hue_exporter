package main

import (
	hue "github.com/collinux/gohue"
	"github.com/mitchellrj/hue_exporter/test"
	"github.com/prometheus/client_golang/prometheus"
	"testing"
)

func TestGroupCollector(t *testing.T) {
	bridge := test.NewStubBridge().WithGroups([]hue.Group{
		hue.Group{
			Action: hue.Action{
				Bri: 256,
				Hue: 100,
				Sat: 80,
			},
			State: struct {
				AllOn bool `json:"all_on"`
				AnyOn bool `json:"any_on"`
			}{
				AllOn: true,
				AnyOn: true,
			},
			Type: "Room",
			Name: "Living room",
		},
	})

	metrics := make(chan prometheus.Metric, 5)
	collector := NewGroupCollector("test_hue", bridge)
	collector.Collect(metrics)
	close(metrics)

	for metric := range metrics {
		t.Logf("%v\n", metric)
	}
}
