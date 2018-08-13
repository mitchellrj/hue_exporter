package main

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/prometheus/common/log"
)

type groupCollector struct {
	bridge             Bridge
	groupBrightness    *prometheus.GaugeVec
	groupHue           *prometheus.GaugeVec
	groupSaturation    *prometheus.GaugeVec
	groupOn            *prometheus.GaugeVec
	groupScrapesFailed prometheus.Counter
}

var variableGroupLabelNames = []string{
	"name",
	"type",
}

// NewGroupCollector Create a new Hue collector for groups
func NewGroupCollector(namespace string, bridge Bridge) prometheus.Collector {
	c := groupCollector{
		bridge: bridge,
		groupBrightness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "group",
				Name:      "brightness",
				Help:      "Group brightness level",
			},
			variableLightLabelNames,
		),
		groupHue: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "group",
				Name:      "hue",
				Help:      "Group hue",
			},
			variableLightLabelNames,
		),
		groupSaturation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "group",
				Name:      "saturation",
				Help:      "Group saturation",
			},
			variableLightLabelNames,
		),
		groupOn: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "group",
				Name:      "on",
				Help:      "Group on  (2 = all group members on, 1 = some group members on, 0 = all group members off)",
			},
			variableLightLabelNames,
		),
		groupScrapesFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "group",
				Name:      "scrapes_failed",
				Help:      "Count of scrapes of group data from the Hue bridge that have failed",
			},
		),
	}

	return c
}

func (c groupCollector) Describe(ch chan<- *prometheus.Desc) {
	c.groupOn.Describe(ch)
	c.groupBrightness.Describe(ch)
	c.groupHue.Describe(ch)
	c.groupSaturation.Describe(ch)
	c.groupScrapesFailed.Describe(ch)
}

func (c groupCollector) Collect(ch chan<- prometheus.Metric) {
	c.groupOn.Reset()
	c.groupBrightness.Reset()
	c.groupHue.Reset()
	c.groupSaturation.Reset()

	groups, err := c.bridge.GetAllGroups()
	if err != nil {
		log.Errorf("Failed to update groups: %v", err)
		c.groupScrapesFailed.Inc()
	}

	for _, group := range groups {
		groupLabels := prometheus.Labels{
			"name":              group.Name,
			"type":              group.Type,
			"model_id":          "",
			"manufacturer_name": "",
			"product_name":      "",
			"unique_id":         "",
		}

		if group.State.AllOn {
			c.groupOn.With(groupLabels).Set(2)
		} else if group.State.AnyOn {
			c.groupOn.With(groupLabels).Set(1)
		} else {
			c.groupOn.With(groupLabels).Set(0)
		}
		c.groupBrightness.With(groupLabels).Set(float64(group.Action.Bri))
		c.groupHue.With(groupLabels).Set(float64(group.Action.Hue))
		c.groupSaturation.With(groupLabels).Set(float64(group.Action.Sat))
	}

	c.groupOn.Collect(ch)
	c.groupBrightness.Collect(ch)
	c.groupHue.Collect(ch)
	c.groupSaturation.Collect(ch)
	c.groupScrapesFailed.Collect(ch)
}
