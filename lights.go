package main

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/prometheus/common/log"
)

type lightCollector struct {
	bridge             Bridge
	lightBrightness    *prometheus.GaugeVec
	lightHue           *prometheus.GaugeVec
	lightSaturation    *prometheus.GaugeVec
	lightOn            *prometheus.GaugeVec
	lightReachable     *prometheus.GaugeVec
	lightScrapesFailed prometheus.Counter
}

var variableLightLabelNames = []string{
	"name",
	"type",
	"model_id",
	"manufacturer_name",
	"product_name",
	"unique_id",
}

// NewLightCollector Create a new Hue collector for lights
func NewLightCollector(namespace string, bridge Bridge) prometheus.Collector {
	c := lightCollector{
		bridge: bridge,
		lightBrightness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "brightness",
				Help:      "Light brightness level",
			},
			variableLightLabelNames,
		),
		lightHue: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "hue",
				Help:      "Light hue",
			},
			variableLightLabelNames,
		),
		lightSaturation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "saturation",
				Help:      "Light saturation",
			},
			variableLightLabelNames,
		),
		lightOn: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "on",
				Help:      "Light on (1 = on, 0 = off)",
			},
			variableLightLabelNames,
		),
		lightReachable: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "reachable",
				Help:      "Light reachability (1/0)",
			},
			variableLightLabelNames,
		),
		lightScrapesFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "light",
				Name:      "scrapes_failed",
				Help:      "Count of scrapes of light data from the Hue bridge that have failed",
			},
		),
	}

	return c
}

func (c lightCollector) Describe(ch chan<- *prometheus.Desc) {
	c.lightOn.Describe(ch)
	c.lightBrightness.Describe(ch)
	c.lightHue.Describe(ch)
	c.lightSaturation.Describe(ch)
	c.lightReachable.Describe(ch)
	c.lightScrapesFailed.Describe(ch)
}

func (c lightCollector) Collect(ch chan<- prometheus.Metric) {
	c.lightOn.Reset()
	c.lightBrightness.Reset()
	c.lightHue.Reset()
	c.lightSaturation.Reset()
	c.lightReachable.Reset()

	lights, err := c.bridge.GetAllLights()
	if err != nil {
		log.Errorf("Failed to update lights: %v", err)
		c.lightScrapesFailed.Inc()
	}

	for _, light := range lights {
		lightLabels := prometheus.Labels{
			"name":              light.Name,
			"type":              light.Type,
			"model_id":          light.ModelID,
			"manufacturer_name": light.ManufacturerName,
			"unique_id":         light.UniqueID,
			"product_name":      light.ProductName,
		}

		if light.State.On {
			c.lightOn.With(lightLabels).Set(1)
		} else {
			c.lightOn.With(lightLabels).Set(0)
		}
		c.lightBrightness.With(lightLabels).Set(float64(light.State.Bri))
		c.lightHue.With(lightLabels).Set(float64(light.State.Hue))
		c.lightSaturation.With(lightLabels).Set(float64(light.State.Saturation))
		if light.State.Reachable {
			c.lightReachable.With(lightLabels).Set(1)
		} else {
			c.lightReachable.With(lightLabels).Set(0)
		}
	}

	c.lightOn.Collect(ch)
	c.lightBrightness.Collect(ch)
	c.lightHue.Collect(ch)
	c.lightSaturation.Collect(ch)
	c.lightReachable.Collect(ch)
	c.lightScrapesFailed.Collect(ch)
}
