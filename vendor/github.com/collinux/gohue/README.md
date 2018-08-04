# GoHue
Package hue interfaces Philips Hue devices to control lights, scenes, schedules, and groups.

[![GoDoc](https://camo.githubusercontent.com/b3b2a2b7fad4e76052830945cd839a3bba5be723/687474703a2f2f696d672e736869656c64732e696f2f62616467652f676f646f632d7265666572656e63652d3532373242342e706e67)](https://godoc.org/github.com/Collinux/GoHue)
[![Go Report Card](https://goreportcard.com/badge/github.com/Collinux/GoHue)](https://goreportcard.com/report/github.com/Collinux/GoHue)

## See GoHue in action!
##### Have a cool project you made using GoHue? Add yours here in a pull request!
[HueBeat](https://github.com/Mobilpadde/HueBeat) by [Mobilpadde](https://github.com/mobilpadde) - Light up a room in sync with your heartbeat.

[BitHue](https://github.com/realytcracker/go-bithue) by [ytcracker](https://github.com/realytcracker) - Light color according to profit gain/loss in bitcoin price

## Installation
```
go get github.com/collinux/gohue
```

## Usage
```
package main

import (
    "github.com/collinux/gohue"
)

func main() {
    // It is recommended that you save the username from bridge.CreateUser
    // so you don't have to press the link button every time and re-auth.
    // When CreateUser is called it will print the generated user token.
    bridgesOnNetwork, _ := hue.FindBridges()
    bridge := bridgesOnNetwork[0]
    username, _ := bridge.CreateUser("someusernamehere")
    bridge.Login(username)

    lights, _ := bridge.GetAllLights()
    for _, light := range lights {
        light.SetBrightness(100)
        light.ColorLoop(true)
    }

    nightstandLight, _ := bridge.GetLightByName("Nightstand")
    nightstandLight.Blink(5)
    nightstandLight.SetName("Bedroom Lamp")

    lights[0].SetColor(hue.RED)
    lights[1].SetColor(hue.BLUE)
    lights[2].SetColor(hue.GREEN)

    for _, light := range lights {
        light.Off()
    }
}
```

## Features
##### Lights
- [x] Get all lights
- [x] Get light by name
- [x] Get light by index on bridge
- [x] Get lights attributes and state
- [x] Set lights attributes (rename)
- [x] Set light state (color, effects, brightness, etc)
- [x] Delete light
- [x] Turn On, Off, Toggle
- [x] Blink
- [x] Colorloop On/Off

##### Bridge
- [x] Create user
- [x] Delete user
- [x] Get configuration
- [ ] Modify configuration
- [ ] Get full state (datastore)
- [x] Search for bridges
- [x] Search for new lights
- [ ] Get all timezones

##### Schedules
- [x] Get all schedules
- [x] Get schedule by ID
- [x] Get schedule attributes
- [ ] Create schedules
- [ ] Set schedule attributes
- [ ] Delete schedule

##### Scenes
- [x] Get all scenes
- [x] Get scene by ID
- [x] Create scene
- [ ] Modify scene
- [ ] Recall scene
- [ ] Delete scene

##### Groups
- [ ] Get all groups
- [ ] Create group
- [ ] Get group attributes
- [ ] Set group attributes
- [ ] Set group state
- [ ] Delete Group

##### Sensors
- [ ] Get all sensors
- [ ] Create sensor
- [ ] Find new sensors
- [ ] Get new sensors
- [ ] Get sensor
- [ ] Update sensor
- [ ] Delete sensor
- [ ] Change sensor configuration

##### Rules
- [ ] Get all rules
- [ ] Get rule
- [ ] Create rule
- [ ] Update rule
- [ ] Delete rule

## API Documentation
This repository is featured on the Philips Hue® developer site and was not developed by "Philips Lighting Holding B.V"...
for official Hue® documentation check out the [Philips Hue® website](http://www.developers.meethue.com/philips-hue-api). This codebase comes with no guaranetees. Use at your own risk.

## License
GoHue - Third party golang library for Philips Hue® gateway interface.
Copyright (C) 2016 Collinux 
GPL version 2 or higher http://www.gnu.org/licenses/gpl.html  

## Contributing  
Pull requests happily accepted on GitHub
