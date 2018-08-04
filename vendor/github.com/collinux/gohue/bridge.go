/*
* bridge.go
* GoHue library for Philips Hue
* Copyright (C) 2016 Collin Guarino (Collinux) collin.guarino@gmail.com
* License: GPL version 2 or higher http://www.gnu.org/licenses/gpl.html
 */
// All things start with the bridge. You will find many Bridge.Func() items
// to use once a bridge has been created and identified.
// See the getting started guide on the Philips hue website:
// http://www.developers.meethue.com/documentation/getting-started

package hue

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Bridge struct defines hardware that is used to communicate with the lights.
type Bridge struct {
	IPAddress string `json:"internalipaddress"`
	Username  string
	Info      BridgeInfo
}

// BridgeInfo struct is the format for parsing xml from a bridge.
type BridgeInfo struct {
	XMLName xml.Name `xml:"root"`
	Device  struct {
		XMLName          xml.Name `xml:"device"`
		DeviceType       string   `xml:"deviceType"`
		FriendlyName     string   `xml:"friendlyName"`
		Manufacturer     string   `xml:"manufacturer"`
		ManufacturerURL  string   `xml:"manufacturerURL"`
		ModelDescription string   `xml:"modelDescription"`
		ModelName        string   `xml:"modelName"`
		ModelNumber      string   `xml:"modelNumber"`
		ModelURL         string   `xml:"modelURL"`
		SerialNumber     string   `xml:"serialNumber"`
		UDN              string   `xml:"UDN"`
	} `xml:"device"`
}

// Get sends an http GET to the bridge
func (bridge *Bridge) Get(path string) ([]byte, io.Reader, error) {
	uri := fmt.Sprintf("http://" + bridge.IPAddress + path)
	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Get(uri)

	if err != nil {
		err = errors.New("unable to access bridge")
		return []byte{}, nil, err
	}
	return HandleResponse(resp)
}

// Put sends an http PUT to the bridge with
// a body formatted with parameters (in a generic interface)
func (bridge *Bridge) Put(path string, params interface{}) ([]byte, io.Reader, error) {
	uri := fmt.Sprintf("http://" + bridge.IPAddress + path)
	client := &http.Client{Timeout: time.Second * 5}

	data, err := json.Marshal(params)
	if err != nil {
		err = errors.New("unable to marshal PUT request interface")
		return []byte{}, nil, err
	}
	//fmt.Println("\n\nPARAMS: ", params)

	request, _ := http.NewRequest("PUT", uri, bytes.NewReader(data))
	resp, err := client.Do(request)
	if err != nil {
		err = errors.New("unable to access bridge")
		return []byte{}, nil, err
	}
	return HandleResponse(resp)
}

// Post sends an http POST to the bridge with
// a body formatted with parameters (in a generic interface).
// If `params` is nil then it will send an empty body with the post request.
func (bridge *Bridge) Post(path string, params interface{}) ([]byte, io.Reader, error) {
	// Add the params to the request or allow an empty body
	request := []byte{}
	if params != nil {
		reqBody, err := json.Marshal(params)
		if err != nil {
			err = errors.New("unable to add POST body parameters due to json marshalling error")
			return []byte{}, nil, err
		}
		request = reqBody
	}
	// Send the request and handle the response
	uri := fmt.Sprintf("http://" + bridge.IPAddress + path)
	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Post(uri, "text/json", bytes.NewReader(request))

	if err != nil {
		err = errors.New("unable to access bridge")
		return []byte{}, nil, err
	}
	return HandleResponse(resp)
}

// Delete sends an http DELETE to the bridge
func (bridge *Bridge) Delete(path string) error {
	uri := fmt.Sprintf("http://" + bridge.IPAddress + path)
	client := &http.Client{Timeout: time.Second * 5}
	req, _ := http.NewRequest("DELETE", uri, nil)
	resp, err := client.Do(req)

	if err != nil {
		err = errors.New("unable to access bridge")
		return err
	}
	_, _, err = HandleResponse(resp)
	return err
}

// HandleResponse manages the http.Response content from a
// bridge Get/Put/Post/Delete by checking it for errors
// and invalid return types.
func HandleResponse(resp *http.Response) ([]byte, io.Reader, error) {
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		trace("Error parsing bridge description xml.", nil)
		return []byte{}, nil, err
	}
	reader := bytes.NewReader(body)
	if strings.Contains(string(body), "\"error\"") {
		errString := string(body)
		errNum := errString[strings.Index(errString, "type\":")+6 : strings.Index(errString, ",\"address")]
		errDesc := errString[strings.Index(errString, "description\":\"")+14 : strings.Index(errString, "\"}}")]
		errOut := fmt.Sprintf("Error type %s: %s.", errNum, errDesc)
		err = errors.New(errOut)
		return []byte{}, nil, err
	}
	return body, reader, nil
}

// FindBridges will visit www.meethue.com/api/nupnp to see a list of
// bridges on the local network.
func FindBridges() ([]Bridge, error) {
	bridge := Bridge{IPAddress: "www.meethue.com"}
	body, _, err := bridge.Get("/api/nupnp")
	if err != nil {
		err = errors.New("unable to locate bridge")
		return []Bridge{}, err
	}
	bridges := []Bridge{}
	err = json.Unmarshal(body, &bridges)
	if err != nil {
		return []Bridge{}, errors.New("unable to parse FindBridges response")
	}
	return bridges, nil
}

// NewBridge defines hardware that is compatible with Hue.
// The function is the core of all functionality, it's necessary
// to call `NewBridge` and `Login` or `CreateUser` to access any
// lights, scenes, groups, etc.
func NewBridge(ip string) (*Bridge, error) {
	bridge := Bridge{
		IPAddress: ip,
	}
	// Test the connection by attempting to get the bridge info.
	err := bridge.GetInfo()
	if err != nil {
		return &Bridge{}, err
	}
	return &bridge, nil
}

// GetInfo retreives the description.xml file from the bridge.
// This is used as a check to see if the bridge is accessible
// and any error will be fatal as the bridge is required for nearly
// all functions.
func (bridge *Bridge) GetInfo() error {
	_, reader, err := bridge.Get("/description.xml")
	if err != nil {
		return err
	}
	data := BridgeInfo{}
	err = xml.NewDecoder(reader).Decode(&data)
	if err != nil {
		err = errors.New("Error: Unable to decode XML response from bridge. ")
		return err
	}
	bridge.Info = data
	return nil
}

// Login verifies that the username token has bridge access
// and only assigns the bridge its Username value if verification is successful.
func (bridge *Bridge) Login(username string) error {
	uri := fmt.Sprintf("/api/%s", username)
	_, _, err := bridge.Get(uri)
	if err != nil {
		return err
	}
	bridge.Username = username
	return nil
}

// CreateUser adds a new user token on the whitelist.
// The token is the first return value in this function which must
// be used with `Bridge.Login`. You cannot use a plaintext username
// like the argument provided in this function.
// This was done by Philips Hue for security reasons.
func (bridge *Bridge) CreateUser(deviceType string) (string, error) {
	params := map[string]string{"devicetype": deviceType}
	body, _, err := bridge.Post("/api", params)
	if err != nil {
		return "", err
	}
	content := string(body)
	username := content[strings.LastIndex(content, ":\"")+2 : strings.LastIndex(content, "\"")]
	bridge.Username = username
	return username, nil
}

// DeleteUser deletes a user given its USER KEY, not the string name.
// See http://www.developers.meethue.com/documentation/configuration-api
// for description on `username` deprecation in place of the devicetype key.
func (bridge *Bridge) DeleteUser(username string) error {
	uri := fmt.Sprintf("/api/%s/config/whitelist/%s", bridge.Username, username)
	err := bridge.Delete(uri)
	if err != nil {
		return err
	}
	return nil
}

// GetAllLights retrieves the state of all lights that the bridge is aware of.
func (bridge *Bridge) GetAllLights() ([]Light, error) {
	uri := fmt.Sprintf("/api/%s/lights", bridge.Username)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return []Light{}, err
	}

	// An index is at the top of every Light in the array
	lightMap := map[string]Light{}
	err = json.Unmarshal(body, &lightMap)
	if err != nil {
		return []Light{}, errors.New("Unable to marshal GetAllLights response. ")
	}

	// Parse the index, add the light to the list, and return the array
	lights := []Light{}
	for index, light := range lightMap {
		light.Index, err = strconv.Atoi(index)
		if err != nil {
			return []Light{}, errors.New("Unable to convert light index to integer. ")
		}
		light.Bridge = bridge
		lights = append(lights, light)
	}
	return lights, nil
}

// GetLightByIndex returns a light struct containing data on
// a light given its index stored on the bridge. This is used for
// quickly updating an individual light.
func (bridge *Bridge) GetLightByIndex(index int) (Light, error) {
	// Send an http GET and inspect the response
	uri := fmt.Sprintf("/api/%s/lights/%d", bridge.Username, index)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return Light{}, err
	}
	if strings.Contains(string(body), "not available") {
		return Light{}, errors.New("Error: Light selection index out of bounds. ")
	}

	// Parse and load the response into the light array
	light := Light{}
	err = json.Unmarshal(body, &light)
	if err != nil {
		return Light{}, errors.New("Error: Unable to unmarshal light data. ")
	}
	light.Index = index
	light.Bridge = bridge
	return light, nil
}

// FindNewLights makes the bridge search the zigbee spectrum for
// lights in the area and will add them to the list of lights available.
// If successful these new lights can be used by `Bridge.GetAllLights`
//
// Notes from Philips Hue API documentation:
// The bridge will search for 1 minute and will add a maximum of 15 new
// lights. To add further lights, the command needs to be sent again after
// the search has completed. If a search is already active, it will be
// aborted and a new search will start.
// http://www.developers.meethue.com/documentation/lights-api#13_search_for_new_lights
func (bridge *Bridge) FindNewLights() error {
	uri := fmt.Sprintf("/api/%s/lights", bridge.Username)
	_, _, err := bridge.Post(uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetLightByName returns a light struct containing data on a given name.
func (bridge *Bridge) GetLightByName(name string) (Light, error) {
	lights, _ := bridge.GetAllLights()
	for _, light := range lights {
		if light.Name == name {
			return light, nil
		}
	}
	errOut := fmt.Sprintf("Error: Light name '%s' not found. ", name)
	return Light{}, errors.New(errOut)
}


// GetAllSensors retrieves the state of all sensors that the bridge is aware of.
func (bridge *Bridge) GetAllSensors() ([]Sensor, error) {
	uri := fmt.Sprintf("/api/%s/sensors", bridge.Username)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return []Sensor{}, err
	}

	// An index is at the top of every sensor in the array
	sensorList := map[string]Sensor{}
	err = json.Unmarshal(body, &sensorList)
	if err != nil {
		fmt.Print(err)
		return []Sensor{}, errors.New("Unable to marshal GetAllSensors response. ")
	}

	// Parse the index, add the sensor to the list, and return the array
	sensors := make([]Sensor, 0, len(sensorList))
	for index, sensor := range sensorList {
		sensor.Index, err = strconv.Atoi(index)
		if err != nil {
			return []Sensor{}, errors.New("Unable to convert sensor index to integer. ")
		}
		sensor.Bridge = bridge
		sensors = append(sensors, sensor)
	}
	return sensors, nil
}

// GetSensorByIndex returns a sensor struct containing data on
// a sensor given its index stored on the bridge.
func (bridge *Bridge) GetSensorByIndex(index int) (Sensor, error) {
	// Send an http GET and inspect the response
	uri := fmt.Sprintf("/api/%s/sensors/%d", bridge.Username, index)
	body, _, err := bridge.Get(uri)
	if err != nil {
		return Sensor{}, err
	}
	if strings.Contains(string(body), "not available") {
		return Sensor{}, errors.New("Error: Sensor selection index out of bounds. ")
	}

	// Parse and load the response into the sensor array
	sensor := Sensor{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		return Sensor{}, errors.New("Error: Unable to unmarshal light data. ")
	}
	sensor.Index = index
	sensor.Bridge = bridge
	return sensor, nil
}

// Log the date, time, file location, line number, and function.
// Message can be "" or Err can be nil (not both)
func trace(message string, err error) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	if err != nil {
		log.Printf("%s:%d %s: %s\n", file, line, f.Name(), err)
	} else {
		log.Printf("%s:%d %s: %s\n", file, line, f.Name(), message)
	}
}
