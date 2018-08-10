package test

import (
	"context"
	"errors"
	hue "github.com/collinux/gohue"
)

type APIFailure int

const (
	LoginFailure APIFailure = iota
	GetGroupsFailure
	GetLightsFailure
	GetSensorsFailure
)

type stubHueBridge struct {
	ctx     context.Context
	lights  []hue.Light
	groups  []hue.Group
	sensors []hue.Sensor
}

func NewStubBridge() *stubHueBridge {
	return &stubHueBridge{
		ctx: context.Background(),
	}
}

func (s *stubHueBridge) WithFailure(fail APIFailure) *stubHueBridge {
	s.ctx = context.WithValue(s.ctx, fail, true)
	return s
}

func (s *stubHueBridge) WithGroups(groups []hue.Group) *stubHueBridge {
	s.groups = groups
	return s
}

func (s *stubHueBridge) WithLights(lights []hue.Light) *stubHueBridge {
	s.lights = lights
	return s
}

func (s *stubHueBridge) WithSensors(sensors []hue.Sensor) *stubHueBridge {
	s.sensors = sensors
	return s
}

func (s *stubHueBridge) Login(apiKey string) error {
	if val, ok := s.ctx.Value(LoginFailure).(bool); ok && val {
		return errors.New("Deliberate login failure")
	}
	return nil
}

func (s *stubHueBridge) GetAllGroups() ([]hue.Group, error) {
	if val, ok := s.ctx.Value(GetGroupsFailure).(bool); ok && val {
		return []hue.Group{}, errors.New("Deliberate get groups failure")
	}
	return s.groups, nil
}

func (s *stubHueBridge) GetAllLights() ([]hue.Light, error) {
	if val, ok := s.ctx.Value(GetLightsFailure).(bool); ok && val {
		return []hue.Light{}, errors.New("Deliberate get lights failure")
	}
	return s.lights, nil
}

func (s *stubHueBridge) GetAllSensors() ([]hue.Sensor, error) {
	if val, ok := s.ctx.Value(GetSensorsFailure).(bool); ok && val {
		return []hue.Sensor{}, errors.New("Deliberate get sensors failure")
	}
	return s.sensors, nil
}
