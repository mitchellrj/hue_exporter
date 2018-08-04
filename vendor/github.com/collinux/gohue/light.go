/*
* light.go
* GoHue library for Philips Hue
* Copyright (C) 2016 Collin Guarino (Collinux) collin.guarino@gmail.com
* License: GPL version 2 or higher http://www.gnu.org/licenses/gpl.html
 */
// http://www.developers.meethue.com/documentation/lights-api

package hue

import (
	"errors"
	"fmt"
	"time"
)

// Light struct defines attributes of a light.
type Light struct {
	State struct {
		On         bool       `json:"on"`        // On or Off state of the light ("true" or "false")
		Bri        uint8      `json:"bri"`       // Brightness value 1-254
		Hue        uint16     `json:"hue"`       // Hue value 1-65535
		Saturation uint8      `json:"sat"`       // Saturation value 0-254
		Effect     string     `json:"effect"`    // "None" or "Colorloop"
		XY         [2]float32 `json:"xy"`        // Coordinates of color in CIE color space
		CT         int        `json:"ct"`        // Mired Color Temperature (google it)
		Alert      string     `json:"alert"`     // "selected" or "none"
		ColorMode  string     `json:"colormode"` // HS or XY mode for choosing color
		Reachable  bool       `json:"reachable"`
	} `json:"state"`
	Type             string  `json:"type"`
	Name             string  `json:"name"`
	ModelID          string  `json:"modelid"`
	ManufacturerName string  `json:"manufacturername"`
	ProductName      string  `json:"productname"`
	UniqueID         string  `json:"uniqueid"`
	SWVersion        string  `json:"swversion"`
	Index            int     // Set by index of light array response
	Bridge           *Bridge // Set by the bridge when the light is found
}

// LightState used in Light.SetState to amend light attributes.
type LightState struct {
	On                  bool        `json:"on"`
	Bri                 uint8       `json:"bri,omitempty"`
	Hue                 uint16      `json:"hue,omitempty"`
	Sat                 uint8       `json:"sat,omitempty"`
	XY                  *[2]float32 `json:"xy,omitempty"`
	CT                  uint16      `json:"ct,omitempty"`
	Effect              string      `json:"effect,omitempty"`
	Alert               string      `json:"alert,omitempty"`
	TransitionTime      string      `json:"transitiontime,omitempty"`
	SaturationIncrement int16       `json:"sat_inc,omitempty"`
	HueIncrement        int32       `json:"hue_inc,omitempty"`
	BrightnessIncrement int16       `json:"bri_inc,omitempty"`
	CTIncrement         int32       `json:"ct_inc,omitempty"`
	XYIncrement         *[2]float32 `json:"xy_inc,omitempty"`
	Name                string      `json:"name,omitempty"`
}

// SetName assigns a new name in the light's
// attributes as recognized by the bridge.
func (light *Light) SetName(name string) error {
	uri := fmt.Sprintf("/api/%s/lights/%d", light.Bridge.Username, light.Index)
	body := make(map[string]string)
	body["name"] = name
	_, _, err := light.Bridge.Put(uri, body)
	if err != nil {
		return err
	}
	return nil
}

// Off turns the light source off
func (light *Light) Off() error {
	return light.SetState(LightState{On: false})
}

// On turns the light source on
func (light *Light) On() error {
	return light.SetState(LightState{On: true})
}

// Toggle switches the light source on and off
func (light *Light) Toggle() error {
	if light.State.On {
		return light.Off()
	}
	return light.On()
}

// Delete removes the light from the
// list of lights available on the bridge.
func (light *Light) Delete() error {
	uri := fmt.Sprintf("/api/%s/lights/%d", light.Bridge.Username, light.Index)
	err := light.Bridge.Delete(uri)
	if err != nil {
		return err
	}
	return nil
}

// Blink increases and decrease the brightness
// repeatedly for a given seconds interval and return the
// light back to its off  or on state afterwards.
// Note: time will vary based on connection speed and algorithm speed.
func (light *Light) Blink(seconds int) error {
	originalPosition := light.State.On
	originalBrightness := light.State.Bri
	blinkMax := 75 // Percent brightness
	blinkMin := 25 // Percent brightness

	// Start with near maximum brightness and toggle between that and
	// a lesser brightness to create a blinking effect.
	for i := 0; i <= seconds*2; i++ {
		if i == 0 {
			err := light.SetBrightness(blinkMax)
			if err != nil {
				return err
			}
		} else if i%2 == 0 {
			err := light.SetBrightness(blinkMax)
			if err != nil {
				return err
			}
		} else {
			err := light.SetBrightness(blinkMin)
			if err != nil {
				return err
			}
		}
		time.Sleep(time.Second / 2)
	}

	// Return the light to its original on or off state and brightness
	if light.State.Bri != originalBrightness || light.State.On != originalPosition {
		light.SetState(LightState{On: originalPosition, Bri: originalBrightness})
	}
	return nil
}

// ColorLoop sets the light state to 'colorloop' if `active`
// is true or it sets the light state to "none" if `activate` is false.
func (light *Light) ColorLoop(activate bool) error {
	var state = "none"
	if activate {
		state = "colorloop"
	}
	return light.SetState(LightState{On: true, Effect: state})
}

// XY positions on the HSL color spectrum used in `Light.SetColor`
// https://en.wikipedia.org/wiki/HSL_and_HSV
var (
	RED    = &[2]float32{0.6915, 0.3083}
	YELLOW = &[2]float32{0.4023, 0.4725}
	ORANGE = &[2]float32{0.4693, 0.4007}
	GREEN  = &[2]float32{0.1700, 0.7000}
	CYAN   = &[2]float32{0.1610, 0.3549}
	BLUE   = &[2]float32{0.1530, 0.0480}
	PURPLE = &[2]float32{0.2363, 0.1154}
	PINK   = &[2]float32{0.3645, 0.1500}
	WHITE  = &[2]float32{0.3227, 0.3290}
)

// SetColor requires a selection from the above light
// color variable section and sets the light to that XY HSL color
func (light *Light) SetColor(color *[2]float32) error {
	lightState := LightState{On: true, XY: color}
	err := light.SetState(lightState)
	if err != nil {
		return err
	}
	return nil
}

// SetColorXY requires a selection from the above light
// color variable section and sets the light to that XY HSL color
// aliased for clarity
func (light *Light) SetColorXY(color *[2]float32) {
	light.SetColor(color)
}

// SetColorHS requires a selection from the above light
// color variable section and sets the light to the Hue value
func (light *Light) SetColorHS(color uint16) error {
	lightState := LightState{On: true, Hue: color}
	err := light.SetState(lightState)
	if err != nil {
		return err
	}
	return nil
}

// Dim lowers the brightness by a percent.
// Note the required value is an integer, for example "20" is converted to 20%.
func (light *Light) Dim(percent int) error {
	if percent > 0 && percent <= 100 {
		originalBri := light.State.Bri
		decreaseBri := float32(originalBri) * float32((float32(percent) / 100.0))
		newBri := uint8(originalBri - uint8(decreaseBri))
		if newBri < 0 {
			newBri = 0
		}
		lightState := LightState{On: true, Bri: newBri}
		err := light.SetState(lightState)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Light.Dim percentage given is not between 1 and 100. ")
}

// SetBrightness sets the brightness to a percentage of the maximum
// maximum brightness as an integer (`LightStruct.Bri between 1 and 254 inclusive`)
func (light *Light) SetBrightness(percent int) error {
	if percent > 0 && percent <= 100 {
		brightness := uint8(float32(percent) * 2.54) // 100=254x --> 2.54
		lightState := LightState{On: true, Bri: brightness}
		err := light.SetState(lightState)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Light.SetBrightness percentage is not between 1 and 100. ")
}

// Brighten will increase LightStruct.Bri by a given percent (integer)
func (light *Light) Brighten(percent int) error {
	if percent > 0 && percent <= 100 {
		originalBri := light.State.Bri
		increaseBri := float32(originalBri) * float32((float32(percent) / 100.0))
		newBri := uint8(originalBri + uint8(increaseBri))
		if newBri > 254 { // LightState.Bri must be between 1 and 254 inclusive
			newBri = 254
		}
		lightState := LightState{On: true, Bri: newBri}
		err := light.SetState(lightState)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Light.Brighten percentage is not between 1 and 100. ")
}

// SetState modifyies light attributes. See `LightState` struct for attributes.
// Brightness must be between 1 and 254 (inclusive)
// Hue must be between 0 and 65535 (inclusive)
// Sat must be between 0 and 254 (inclusive)
// See http://www.developers.meethue.com/documentation/lights-api for more info
func (light *Light) SetState(newState LightState) error {
	uri := fmt.Sprintf("/api/%s/lights/%d/state", light.Bridge.Username, light.Index)
	_, _, err := light.Bridge.Put(uri, newState)
	if err != nil {
		return err
	}

	// Get the new light state and update the current Light struct
	*light, err = light.Bridge.GetLightByIndex(light.Index)
	if err != nil {
		return err
	}
	return nil
}
