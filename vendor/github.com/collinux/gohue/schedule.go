/*
* schedule.go
* GoHue library for Philips Hue
* Copyright (C) 2016 Collin Guarino (Collinux) collin.guarino@gmail.com
* License: GPL version 2 or higher http://www.gnu.org/licenses/gpl.html
 */
// http://www.developers.meethue.com/documentation/schedules-api-0

package hue

import (
	"encoding/json"
	"fmt"
)

// Schedule struct defines attributes of Alarms and Timers
type Schedule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     struct {
		Address string `json:"address"`
		Body    struct {
			Scene string `json:"scene"`
		} `json:"body"`
		Method string `json:"method"`
	} `json:"command"`
	Localtime  string `json:"localtime"`
	Time       string `json:"time"`
	Created    string `json:"created"`
	Status     string `json:"status"`
	Autodelete bool   `json:"autodelete"`
	ID         string
}

// GetAllSchedules gets Alarms and Timers in a Schedule struct.
func (bridge *Bridge) GetAllSchedules() ([]Schedule, error) {
	uri := fmt.Sprintf("/api/%s/schedules", bridge.Username)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return []Schedule{}, err
	}

	// Each index key is the topmost element of the json array.
	// Unmarshal the array, loop through each index key, and add it to the list
	schedules := map[string]Schedule{}
	err = json.Unmarshal(body, &schedules)
	if err != nil {
		return []Schedule{}, err
	}
	scheduleList := []Schedule{}
	for key, value := range schedules {
		schedule := Schedule{}
		schedule = value
		schedule.ID = key
		scheduleList = append(scheduleList, schedule)
	}
	return scheduleList, nil
}

// GetSchedule gets the attributes for an individual schedule.
// This is used to optimize time when updating the state of a schedule item.
// Note: The ID is not an index, it's a unique key generated for each schedule.
func (bridge *Bridge) GetSchedule(id string) (Schedule, error) {
	uri := fmt.Sprintf("/api/%s/schedules/%s", bridge.Username, id)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return Schedule{}, err
	}

	schedule := Schedule{}
	err = json.Unmarshal(body, &schedule)
	if err != nil {
		return Schedule{}, err
	}
	return schedule, nil
}

// CreateSchedule TODO: NOT TESTED, NOT FULLY IMPLEMENTED
func (bridge *Bridge) CreateSchedule(schedule Schedule) error {
	uri := fmt.Sprintf("/api/%s/schedules", bridge.Username)
	body, _, err := bridge.Post(uri, schedule)
	if err != nil {
		return err
	}

	fmt.Println("CREATE SCHEDULE BODY: ", string(body))
	return nil
}

// func (schedule *Schedule) Disable() {
//
// }
//
// func (schedule *Schedule) Enable() {
//
// }
//
//
// func (bridge *Bridge) GetSchedule(index int) (interface{}, error) {
//     return []interface{}, nil
// }
//
// func (bridge *Bridge) SetSchedule(index int, schedule interface{}) error {
//     return nil
// }
//
// func (bridge *Bridge) DeleteSchedule(index int) error {
//     return nil
// }
