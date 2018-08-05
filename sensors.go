package main

import (
	hue "github.com/collinux/gohue"
	log "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type sensorCollector struct {
	bridge              *hue.Bridge
	ignoreTypes         []string
	sensorValue         *prometheus.GaugeVec
	sensorLastUpdated   *prometheus.GaugeVec
	sensorOn            *prometheus.GaugeVec
	sensorBattery       *prometheus.GaugeVec
	sensorReachable     *prometheus.GaugeVec
	sensorScrapesFailed prometheus.Counter
}

var variableSensorLabelNames = []string{
	"name",
	"type",
	"model_id",
	"manufacturer_name",
	"product_name",
	"unique_id",
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// NewSensorCollector Create a new Hue collector for sensors
func NewSensorCollector(namespace string, bridge *hue.Bridge, ignoreTypes []string) prometheus.Collector {
	c := sensorCollector{
		bridge:      bridge,
		ignoreTypes: ignoreTypes,
		sensorValue: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "sensor",
				Name:      "value",
				Help:      "Sensor values",
			},
			variableSensorLabelNames,
		),
		sensorBattery: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "sensor",
				Name:      "battery",
				Help:      "Sensor battery levels (%)",
			},
			variableSensorLabelNames,
		),
		sensorLastUpdated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "sensor",
				Name:      "last_updated",
				Help:      "Sensor last updated time",
			},
			variableSensorLabelNames,
		),
		sensorOn: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "sensor",
				Name:      "on",
				Help:      "Sensor on/off (1/0)",
			},
			variableSensorLabelNames,
		),
		sensorReachable: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "sensor",
				Name:      "reachable",
				Help:      "Sensor reachability (1/0)",
			},
			variableSensorLabelNames,
		),
		sensorScrapesFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "sensor",
				Name:      "scrapes_failed",
				Help:      "Count of scrapes of sensor data from the Hue bridge that have failed",
			},
		),
	}

	return c
}

func (c sensorCollector) Describe(ch chan<- *prometheus.Desc) {
	c.sensorValue.Describe(ch)
	c.sensorBattery.Describe(ch)
	c.sensorLastUpdated.Describe(ch)
	c.sensorOn.Describe(ch)
	c.sensorReachable.Describe(ch)
	c.sensorScrapesFailed.Describe(ch)
}

func (c sensorCollector) Collect(ch chan<- prometheus.Metric) {
	c.sensorValue.Reset()
	c.sensorBattery.Reset()
	c.sensorLastUpdated.Reset()
	c.sensorOn.Reset()
	c.sensorReachable.Reset()

	sensors, err := c.bridge.GetAllSensors()
	if err != nil {
		log.Errorf("Failed to update sensors: %v", err)
		c.sensorScrapesFailed.Inc()
	}

	for _, sensor := range sensors {
		var sensorValue float64
		if contains(c.ignoreTypes, sensor.Type) {
			continue
		} else if sensor.Type == "Daylight" {
			// bridge daylight (sunrise / sunset) sensor
			if sensor.State.Daylight {
				sensorValue = 1
			}
		} else if sensor.Type == "ZGPSwitch" {
			// Hue tap switch
			sensorValue = float64(sensor.State.ButtonEvent)
		} else if sensor.Type == "ZLLSwitch" {
			// Hue dimmer switch
			sensorValue = float64(sensor.State.ButtonEvent)
		} else if sensor.Type == "ClipGenericStatus" {
			sensorValue = float64(sensor.State.Status)
		} else if sensor.Type == "ZLLTemperature" {
			sensorValue = float64(sensor.State.Temperature)
		} else if sensor.Type == "ZLLPresence" {
			if sensor.State.Presence {
				sensorValue = 1
			}
		} else if sensor.Type == "ZLLLightLevel" {
			sensorValue = float64(sensor.State.LightLevel)
		}

		sensorLabels := prometheus.Labels{
			"name":              sensor.Name,
			"model_id":          sensor.ModelID,
			"manufacturer_name": sensor.ManufacturerName,
			"type":              sensor.Type,
			"unique_id":         sensor.UniqueID,
			"product_name":      sensor.ProductName,
		}

		c.sensorValue.With(sensorLabels).Set(sensorValue)
		c.sensorBattery.With(sensorLabels).Set(float64(sensor.Config.Battery))
		c.sensorLastUpdated.With(sensorLabels).Set(float64(sensor.State.LastUpdated.Unix()))
		if sensor.Config.On {
			c.sensorOn.With(sensorLabels).Set(1)
		} else {
			c.sensorOn.With(sensorLabels).Set(0)
		}
		if sensor.Config.Reachable {
			c.sensorReachable.With(sensorLabels).Set(1)
		} else {
			c.sensorReachable.With(sensorLabels).Set(0)
		}
	}

	c.sensorValue.Collect(ch)
	c.sensorBattery.Collect(ch)
	c.sensorLastUpdated.Collect(ch)
	c.sensorOn.Collect(ch)
	c.sensorReachable.Collect(ch)
	c.sensorScrapesFailed.Collect(ch)
}
