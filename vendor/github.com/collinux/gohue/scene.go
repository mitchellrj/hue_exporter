/*
* scene.go
* GoHue library for Philips Hue
* Copyright (C) 2016 Collin Guarino (Collinux) collin.guarino@gmail.com
* License: GPL version 2 or higher http://www.gnu.org/licenses/gpl.html
 */
// http://www.developers.meethue.com/documentation/scenes-api

package hue

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Scene struct defines attributes for Scene items
type Scene struct {
	Appdata *struct {
		Data    string `json:"data,omitempty"`
		Version int    `json:"version,omitempty"`
	} `json:"appdata,omitempty"`
	Lastupdated string   `json:"lastupdated,omitempty"`
	Lights      []string `json:"lights,omitempty"`
	Locked      bool     `json:"locked,omitempty"`
	Name        string   `json:"name,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	Picture     string   `json:"picture,omitempty"`
	Recycle     bool     `json:"recycle,omitempty"`
	Version     int      `json:"version,omitempty"`
	ID          string   `json:",omitempty"`
}

// GetAllScenes gets the attributes for all scenes.
func (bridge *Bridge) GetAllScenes() ([]Scene, error) {
	uri := fmt.Sprintf("/api/%s/scenes", bridge.Username)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return []Scene{}, err
	}

	scenes := map[string]Scene{}
	err = json.Unmarshal(body, &scenes)
	if err != nil {
		return []Scene{}, err
	}
	scenesList := []Scene{}
	for key, value := range scenes {
		scene := Scene{}
		scene = value
		scene.ID = key
		scenesList = append(scenesList, scene)
	}
	return scenesList, nil
}

// GetScene gets the attributes for an individual scene.
// This is used to optimize time when updating the state of the scene.
// Note: The ID is not an index, it's a unique key generated for each scene.
func (bridge *Bridge) GetScene(id string) (Scene, error) {
	uri := fmt.Sprintf("/api/%s/scenes/%s", bridge.Username, id)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return Scene{}, err
	}

	scene := Scene{}
	err = json.Unmarshal(body, &scene)
	if err != nil {
		return Scene{}, err
	}
	return scene, nil
}

// GetSceneByName gets the attributes for the scene identified by a name
func (bridge *Bridge) GetSceneByName(name string) (Scene, error) {

	scenes, _ := bridge.GetAllScenes()

	// Iterate in reverse, as later entries seem to be the newest
	for i := len(scenes) - 1; i >= 0; i-- {
		if scenes[i].Name == name {
			return scenes[i], nil
		}
	}

	errOut := fmt.Sprintf("Error: Scene name '%s' not found. ", name)
	return Scene{}, errors.New(errOut)
}

// RecallScene recalls a scene
func (bridge *Bridge) RecallScene(id string) error {
	action := &Action{Scene: id}
	return bridge.SetGroupState(0, action)
}

// RecallSceneByName recalls a scene
func (bridge *Bridge) RecallSceneByName(name string) error {
	scene, err := bridge.GetSceneByName(name)
	if err != nil {
		return err
	}
	return bridge.RecallScene(scene.ID)
}

// CreateScene posts a new scene configuration to the bridge.
func (bridge *Bridge) CreateScene(scene Scene) error {
	uri := fmt.Sprintf("/api/%s/scenes/", bridge.Username)
	_, _, err := bridge.Post(uri, scene)
	if err != nil {
		return err
	}
	return nil
}

// Bridge.ModifySceneState amends light states for lights
// included in a scene list. See `Bridge.ModifyScene` for
// changing the lights included in the scene list.
// func (bridge *Bridge) ModifySceneState() error {
//
// }

// Bridge.ModifyScene amends the lights included for a given scene or
// it can be used to change the scene name. To amend light states for
// lights included in a scene list see `Bridge.ModifySceneState`.
// func (bridge *Bridge) ModifyScene() error {
//     uri := fmt.Sprintf("/api/%s/scenes/%s/lightstates/%s",
//         bridge.Username, oldScene.ID, )
// }
