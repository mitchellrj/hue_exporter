/*
* group.go
* GoHue library for Philips Hue
* Copyright (C) 2016 Collin Guarino (Collinux) collin.guarino@gmail.com
* License: GPL version 2 or higher http://www.gnu.org/licenses/gpl.html
 */
// http://www.developers.meethue.com/documentation/groups-api

package hue

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Action struct defines the state of a group
type Action struct {
	Alert     string    `json:"alert,omitempty"`
	Bri       int       `json:"bri,omitempty"`
	Colormode string    `json:"colormode,omitempty"`
	Ct        int       `json:"ct,omitempty"`
	Effect    string    `json:"effect,omitempty"`
	Hue       int       `json:"hue,omitempty"`
	On        bool      `json:"on,omitempty"`
	Sat       int       `json:"sat,omitempty"`
	XY        []float64 `json:"xy,omitempty"`
	Scene     string    `json:"scene,omitempty"`
}

// Group struct defines the attributes for a group of lights.
type Group struct {
	Action Action `json:"action"`
	State  struct {
		AllOn bool `json:"all_on"`
		AnyOn bool `json:"any_on"`
	} `json:"state"`

	Recycle bool     `json:"recycle"`
	Class   string   `json:"class"`
	Lights  []string `json:"lights"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Bridge  *Bridge
	Index   int
}

// GetAllGroups gets the attributes for each group of lights.
func (bridge *Bridge) GetAllGroups() ([]Group, error) {
	uri := fmt.Sprintf("/api/%s/groups", bridge.Username)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return []Group{}, err
	}

	groupList := map[string]Group{}
	err = json.Unmarshal(body, &groupList)
	if err != nil {
		return []Group{}, err
	}

	// Parse the index, add the sensor to the list, and return the array
	groups := make([]Group, 0, len(groupList))
	for index, group := range groupList {
		group.Index, err = strconv.Atoi(index)
		if err != nil {
			return []Group{}, errors.New("Unable to convert group index to integer. ")
		}
		group.Bridge = bridge
		groups = append(groups, group)
	}
	return groups, nil
}

// SetGroupState sends an action to group
func (bridge *Bridge) SetGroupState(group int, action *Action) error {
	uri := fmt.Sprintf("/api/%s/groups/%d/action", bridge.Username, group)
	_, _, err := bridge.Put(uri, action)
	if err != nil {
		return err
	}
	return nil
}
