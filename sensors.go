package main

import (
	"math"

	hue "github.com/collinux/gohue"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/prometheus/common/log"
)

type sensorCollector struct {
	bridge              Bridge
	ignoreTypes         []string
	matchNames          bool
	sensorValue         *prometheus.GaugeVec
	sensorLastUpdated   *prometheus.GaugeVec
	sensorOn            *prometheus.GaugeVec
	sensorBattery       *prometheus.GaugeVec
	sensorReachable     *prometheus.GaugeVec
	sensorScrapesFailed prometheus.Counter
	bridgeRestarts      prometheus.Counter
}

var variableSensorLabelNames = []string{
	"name",
	"type",
	"model_id",
	"manufacturer_name",
	"product_name",
	"unique_id",
	"device_id",
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// NewSensorCollector Create a new Hue collector for sensors
func NewSensorCollector(namespace string, bridge Bridge, ignoreTypes []string, matchNames bool) prometheus.Collector {
	c := sensorCollector{
		bridge:      bridge,
		ignoreTypes: ignoreTypes,
		matchNames:  matchNames,
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
		bridgeRestarts: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "bridge",
				Name:      "restarts",
				Help:      "Count of number of bridge restarts detected",
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
	c.bridgeRestarts.Describe(ch)
}

func (c sensorCollector) recordSensor(sensor hue.Sensor, sensorName string, deviceID string, sensorValue float64) {
	sensorLabels := prometheus.Labels{
		"name":              sensorName,
		"model_id":          sensor.ModelID,
		"manufacturer_name": sensor.ManufacturerName,
		"type":              sensor.Type,
		"unique_id":         sensor.UniqueID,
		"device_id":         deviceID,
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
	sensorNames := make(map[string]string)
	sensorLastUpdatedHistory := make(map[string]int64)
	restartDetected := false

	for _, sensor := range sensors {
		var sensorValue float64
		deviceID := sensor.UniqueID
		if contains(c.ignoreTypes, sensor.Type) {
			continue
		} else if sensor.Type == "Daylight" {
			// bridge daylight (sunrise / sunset) sensor
			if sensor.State.Daylight {
				sensorValue = 1
			}
		} else if sensor.Type == "ZGPSwitch" {
			// Hue tap switch
			deviceID = sensor.UniqueID[0:23]
			sensorValue = float64(sensor.State.ButtonEvent)
		} else if sensor.Type == "ZLLSwitch" {
			// Hue dimmer switch
			deviceID = sensor.UniqueID[0:23]
			sensorValue = float64(sensor.State.ButtonEvent)
		} else if sensor.Type == "ClipGenericStatus" {
			sensorValue = float64(sensor.State.Status)
		} else if sensor.Type == "ZLLPresence" {
			deviceID = sensor.UniqueID[0:23]
			sensorNames[deviceID] = sensor.Name
			if sensor.State.Presence {
				sensorValue = 1
			}
		} else {
			continue
		}
		if sensorLastUpdatedHistory[sensor.UniqueID] > math.MinInt64 && sensor.State.LastUpdated.Unix() == math.MinInt64 {
			restartDetected = true
		}
		sensorLastUpdatedHistory[sensor.UniqueID] = sensor.State.LastUpdated.Unix()
		c.recordSensor(sensor, sensor.Name, deviceID, sensorValue)
	}
	// kinda inefficient looping over them twice, but simplies code when name matching is enabled
	for _, sensor := range sensors {
		var sensorValue float64
		if sensor.Type == "ZLLTemperature" {
			sensorValue = float64(sensor.State.Temperature)
		} else if sensor.Type == "ZLLLightLevel" {
			sensorValue = float64(sensor.State.LightLevel)
		} else {
			continue
		}
		deviceID := sensor.UniqueID[0:23]
		sensorName := sensor.Name
		if c.matchNames {
			var ok bool
			sensorName, ok = sensorNames[deviceID]
			if !ok {
				sensorName = sensor.Name
			}
		}
		if sensorLastUpdatedHistory[sensor.UniqueID] > math.MinInt64 && sensor.State.LastUpdated.Unix() == math.MinInt64 {
			restartDetected = true
		}
		sensorLastUpdatedHistory[sensor.UniqueID] = sensor.State.LastUpdated.Unix()
		c.recordSensor(sensor, sensorName, deviceID, sensorValue)
	}

	if restartDetected {
		c.bridgeRestarts.Inc()
	}

	c.sensorValue.Collect(ch)
	c.sensorBattery.Collect(ch)
	c.sensorLastUpdated.Collect(ch)
	c.sensorOn.Collect(ch)
	c.sensorReachable.Collect(ch)
	c.sensorScrapesFailed.Collect(ch)
	c.bridgeRestarts.Collect(ch)
}
