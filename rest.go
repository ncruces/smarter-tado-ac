package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (ctx *TadoContext) tadoGet(url string) ([]byte, error) {
	res, err := ctx.tado.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(http.StatusText(res.StatusCode))
	}

	return ioutil.ReadAll(res.Body)
}

func (ctx *TadoContext) getTadoAccount() (ret TadoAccount, err error) {
	buf, err := ctx.tadoGet("https://my.tado.com/api/v2/me")

	if err != nil {
		err = fmt.Errorf("GET /api/v2/me: %w", err)
		return
	}

	if err = json.Unmarshal(buf, &ret); err != nil {
		err = fmt.Errorf("decode /api/v2/me: %w", err)
		return
	}

	return
}

func (ctx *TadoContext) getTadoHome(id int) (ret TadoHome, err error) {
	buf, err := ctx.tadoGet("https://my.tado.com/api/v2/homes/" + strconv.Itoa(id))

	if err != nil {
		err = fmt.Errorf("GET /api/v2/homes/[home]: %w", err)
		return
	}

	if err = json.Unmarshal(buf, &ret); err != nil {
		err = fmt.Errorf("decode /api/v2/homes/[home]: %w", err)
		return
	}

	return
}

func (ctx *TadoContext) getTadoZones(home TadoHome) (ret []TadoZone, err error) {
	buf, err := ctx.tadoGet("https://my.tado.com/api/v2/homes/" + strconv.Itoa(home.ID) + "/zones")

	if err != nil {
		err = fmt.Errorf("GET /api/v2/homes/[home]/zones: %w", err)
		return
	}

	if err = json.Unmarshal(buf, &ret); err != nil {
		err = fmt.Errorf("decode /api/v2/homes/[home]/zones: %w", err)
		return
	}

	return
}

func (ctx *TadoContext) getTadoZoneState(home TadoHome, zone TadoZone) (ret TadoZoneState, err error) {
	buf, err := ctx.tadoGet("https://my.tado.com/api/v2/homes/" + strconv.Itoa(home.ID) + "/zones/" + strconv.Itoa(zone.ID) + "/state")

	if err != nil {
		err = fmt.Errorf("GET /api/v2/homes/[home]/zones/[zone]/state: %w", err)
		return
	}

	if err = json.Unmarshal(buf, &ret); err != nil {
		err = fmt.Errorf("decode /api/v2/homes/[home]/zones/[zone]/state: %w", err)
		return
	}

	return
}

func (ctx *TadoContext) getTadoAwayConfiguration(home TadoHome, zone TadoZone) (ret TadoAwayConfiguration, err error) {
	buf, err := ctx.tadoGet("https://my.tado.com/api/v2/homes/" + strconv.Itoa(home.ID) + "/zones/" + strconv.Itoa(zone.ID) + "/awayConfiguration")

	if err != nil {
		err = fmt.Errorf("GET /api/v2/homes/[home]/zones/[zone]/awayConfiguration: %w", err)
		return
	}

	if err = json.Unmarshal(buf, &ret); err != nil {
		err = fmt.Errorf("decode /api/v2/homes/[home]/zones/[zone]/awayConfiguration: %w", err)
		return
	}

	return
}

func (ctx *TadoContext) getTadoActiveTimetable(home TadoHome, zone TadoZone) (ret TadoActiveTimetable, err error) {
	buf, err := ctx.tadoGet("https://my.tado.com/api/v2/homes/" + strconv.Itoa(home.ID) + "/zones/" + strconv.Itoa(zone.ID) + "/schedule/activeTimetable")

	if err != nil {
		err = fmt.Errorf("GET /api/v2/homes/[home]/zones/[zone]/schedule/activeTimetable: %w", err)
		return
	}

	if err = json.Unmarshal(buf, &ret); err != nil {
		err = fmt.Errorf("decode /api/v2/homes/[home]/zones/[zone]/schedule/activeTimetable: %w", err)
		return
	}

	return
}

func (ctx *TadoContext) getTadoTimetableBlock(home TadoHome, zone TadoZone, timetable TadoActiveTimetable, instant time.Time) (ret TadoTimetableBlock, err error) {
	loc, err := time.LoadLocation(home.DateTimeZone)
	if err != nil {
		err = fmt.Errorf("invalid home.DateTimeZone: %s", home.DateTimeZone)
		return
	}

	instant = instant.In(loc)
	day := TadoDayTypeMap[timetable.ID][instant.Weekday()]
	buf, err := ctx.tadoGet("https://my.tado.com/api/v2/homes/" + strconv.Itoa(home.ID) + "/zones/" + strconv.Itoa(zone.ID) + "/schedule/timetables/" + strconv.Itoa(timetable.ID) + "/blocks/" + day)

	if err != nil {
		err = fmt.Errorf("GET /api/v2/homes/[home]/zones/[zone]/schedule/timetables/[timetable]/blocks/[day]: %w", err)
		return
	}

	var blocks []TadoTimetableBlock
	if err = json.Unmarshal(buf, &blocks); err != nil {
		err = fmt.Errorf("decode /api/v2/homes/[home]/zones/[zone]/schedule/timetables/[timetable]/blocks/[day]: %w", err)
		return
	}

	for _, block := range blocks {
		start, _ := time.Parse("15:04", block.Start)
		end, _ := time.Parse("15:04", block.End)
		start = time.Date(instant.Year(), instant.Month(), instant.Day(), start.Hour(), start.Minute(), 0, 0, loc)
		end = time.Date(instant.Year(), instant.Month(), instant.Day(), end.Hour(), end.Minute(), 0, 0, loc)
		if start.After(end) {
			end = end.AddDate(0, 0, 1)
		}
		if !start.After(instant) && instant.Before(end) {
			return block, nil
		}
	}

	return
}

func (ctx *TadoContext) putTadoOverlay(home TadoHome, zone TadoZone, overlay TadoOverlay) (TadoOverlay, error) {
	var ret TadoOverlay

	buf, err := json.Marshal(overlay)
	if err != nil {
		err = fmt.Errorf("encode overlay: %w", err)
		return ret, err
	}

	req, err := http.NewRequest("PUT", "https://my.tado.com/api/v2/homes/"+strconv.Itoa(home.ID)+"/zones/"+strconv.Itoa(zone.ID)+"/overlay", bytes.NewReader(buf))
	if err != nil {
		err = fmt.Errorf("PUT /api/v2/homes/[home]/zones/[zone]/overlay: %w", err)
		return ret, err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := ctx.tado.Do(req)

	if err != nil {
		err = fmt.Errorf("PUT /api/v2/homes/[home]/zones/[zone]/overlay: %w", err)
		return ret, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("PUT /api/v2/homes/[home]/zones/[zone]/overlay: %s", http.StatusText(res.StatusCode))
		return ret, err
	}

	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		err = fmt.Errorf("decode /api/v2/homes/[home]/zones/[zone]/overlay: %w", err)
		return ret, err
	}

	return ret, nil
}
