/*
* sensor.go
* GoHue library for Philips Hue
* Copyright (C) 2016 Collin Guarino (Collinux) collin.guarino@gmail.com
* License: GPL version 2 or higher http://www.gnu.org/licenses/gpl.html
 */
// https://developers.meethue.com/documentation/sensors-api

package hue

import (
	"strings"
	"time"
)

// special time type for unmarshal of lastupdated
type UpdateTime struct {
	*time.Time
}

// implement Unmarshal interface
// required for "none" as lastupdated in unused sensor
func (u *UpdateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "none" || s == "null" {
		*u = UpdateTime{&time.Time{}}
		return nil
	}
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return err
	}
	*u = UpdateTime{&t}
	return nil
}

// Sensor struct defines attributes of a sensor.
type Sensor struct {
	State struct {
		Daylight    bool       `json:"daylight"`    // True if day & false if night
		LastUpdated UpdateTime `json:"lastupdated"` // Time of last update
		ButtonEvent uint16     `json:"buttonevent"` // ID of button event
		Status      uint16     `json:"status"`      // Generic status value
		Temperature int16      `json:"temperature"` // Temperature in hundredths °C
		LightLevel  uint16     `json:"lightlevel"`  // Light level in Lux
		Dark        bool       `json:"dark"`        // True if sensor configuration judges the current light level to be "dark"
		Presence    bool       `json:"presence"`    // True if the sensor has detected movement
	} `json:"state"`

	Config struct {
		On              bool   `json:"on"`             // Turns the sensor on/off. When off, state changes of the sensor are not reflected in the sensor resource.
		Reachable       bool   `json:"reachable"`      // Indicates whether communication with devices is possible
		Battery         uint8  `json:"battery"`        // The current battery state in percent, only for battery powered devices
		SunriseOffset   int8   `json:"sunriseoffset"`  // Offset from sunrise in minutes after which the sensor considers it to be "daylight"
		SunsetOffset    int8   `json:"sunsetoffset"`   // Offset from sunset in minutes before which the sensor considers it to be "daylight"
		Alert           string `json:"alert"`          // Motion sensor "alert" state - usually "none"
		LEDIndication   bool   `json:"ledindication"`  // True if the motion sensor's LED is enabled
		UserTest        bool   `json:"usertest"`       // True if the motion sensor is in user test mode
		Sensitivity     uint8  `json:"sensitivity"`    // Motion sensor movement sensitivity level
		SensitivityMax  uint8  `json:"sensitivitymax"` // Motion sensor maximum movement sensitivity level
		ThresholdDark   uint16 `json:"tholddark"`      // Motion sensor threshold below which it is considered "dark" (Lux)
		ThresholdOffset uint16 `json:"tholdoffset"`    // Motion sensor level above darkness threshold that is considered "daylight" (Lux)
	} `json:"config"`

	SwUpdate struct {
		State       string     `json:"state"`       // Usually "noupdates"
		LastInstall UpdateTime `json:"lastinstall"` // Last time software was updated on this device
	} `json:"swupdate"`

	Capabilities struct {
		Certified bool `json:"certified"` // Certified Hue product
	} `json:"capabilities"`

	Type             string  `json:"type"`
	Name             string  `json:"name"`
	ModelID          string  `json:"modelid"`
	ManufacturerName string  `json:"manufacturername"`
	ProductName      string  `json:"productname"`
	UniqueID         string  `json:"uniqueid"`
	SWVersion        string  `json:"swversion"`
	Recycle          bool    `json:"recycle"`
	Index            int     // Set by index of sensor array response
	Bridge           *Bridge // Set by the bridge when the sensor is found
}

/// Refresh sensor attributes
func (s *Sensor) Refresh() error {
	sensor, err := s.Bridge.GetSensorByIndex(s.Index)
	if err != nil {
		return err
	}

	s.State = sensor.State
	s.Config = sensor.Config
	s.SWVersion = sensor.SWVersion
	s.Name = sensor.Name
	return nil
}
