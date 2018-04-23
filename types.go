package main

import (
	"time"
)

var TadoDayTypeMap = [3][7]string{
	{"MONDAY_TO_SUNDAY", "MONDAY_TO_SUNDAY", "MONDAY_TO_SUNDAY", "MONDAY_TO_SUNDAY", "MONDAY_TO_SUNDAY", "MONDAY_TO_SUNDAY", "MONDAY_TO_SUNDAY"},
	{"SUNDAY", "MONDAY_TO_FRIDAY", "MONDAY_TO_FRIDAY", "MONDAY_TO_FRIDAY", "MONDAY_TO_FRIDAY", "MONDAY_TO_FRIDAY", "SATURDAY"},
	{"SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY"},
}

type TadoAccount struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Homes []TadoHome `json:"homes"`
}

type TadoHome struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	DateTimeZone    string `json:"dateTimeZone"`
	TemperatureUnit string `json:"temperatureUnit"`
}

type TadoZone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type TadoZoneState struct {
	TadoMode         string      `json:"tadoMode"`
	Setting          TadoSetting `json:"setting"`
	Overlay          TadoOverlay `json:"overlay"`
	SensorDataPoints struct {
		InsideTemperature struct {
			Celsius    float64   `json:"celsius"`
			Fahrenheit float64   `json:"fahrenheit"`
			Timestamp  time.Time `json:"timestamp"`
			Type       string    `json:"type"`
			Precision  struct {
				Celsius    float64 `json:"celsius"`
				Fahrenheit float64 `json:"fahrenheit"`
			} `json:"precision"`
		} `json:"insideTemperature"`
		Humidity struct {
			Type       string    `json:"type"`
			Percentage float64   `json:"percentage"`
			Timestamp  time.Time `json:"timestamp"`
		} `json:"humidity"`
	} `json:"sensorDataPoints"`
}

type TadoActiveTimetable struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type TadoTimetableBlock struct {
	DayType             string      `json:"dayType"`
	Start               string      `json:"start"`
	End                 string      `json:"end"`
	GeolocationOverride bool        `json:"geolocationOverride"`
	Setting             TadoSetting `json:"setting"`
}

type TadoTemperature struct {
	Celsius    float64 `json:"celsius,omitempty"`
	Fahrenheit float64 `json:"fahrenheit,omitempty"`
}

type TadoOverlay struct {
	Type        string          `json:"type,omitempty"`
	Setting     TadoSetting     `json:"setting,omitempty"`
	Termination TadoTermination `json:"termination,omitempty"`
}

type TadoAwayConfiguration struct {
	Type    string      `json:"type,omitempty"`
	Setting TadoSetting `json:"setting,omitempty"`
}

type TadoSetting struct {
	Type        string           `json:"type,omitempty"`
	Power       string           `json:"power,omitempty"`
	Mode        string           `json:"mode,omitempty"`
	FanSpeed    string           `json:"fanSpeed,omitempty"`
	Temperature *TadoTemperature `json:"temperature,omitempty"`
}

type TadoTermination struct {
	Type                   string     `json:"type,omitempty"`
	DurationInSeconds      int        `json:"durationInSeconds,omitempty"`
	RemainingTimeInSeconds int        `json:"remainingTimeInSeconds,omitempty"`
	Expiry                 *time.Time `json:"expiry,omitempty"`
	ProjectedExpiry        *time.Time `json:"projectedExpiry,omitempty"`
}
