package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"os"
	"time"
)

type TadoContext struct {
	tado *http.Client
	time time.Time
}

type TadoConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	ctx := makeContext()

	acc, err := ctx.getTadoAccount()
	if err != nil {
		panic(err)
	}

	for _, home := range acc.Homes {
		home, err = ctx.getTadoHome(home.ID)
		if err != nil {
			panic(err)
		}

		zones, err := ctx.getTadoZones(home)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s...\n", home.Name)
		for _, zone := range zones {
			if zone.Type == "AIR_CONDITIONING" {
				fmt.Printf("%s, ", zone.Name)
				err = ctx.smartZone(home, zone)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (ctx *TadoContext) smartZone(home TadoHome, zone TadoZone) error {
	state, err := ctx.getTadoZoneState(home, zone)
	if err != nil {
		return err
	}

	if expiry := state.Overlay.Termination.ProjectedExpiry; expiry != nil && expiry.After(ctx.time.Add(10*time.Minute)) {
		fmt.Print("manual mode\n")
		return nil
	}

	timetable, err := ctx.getTadoActiveTimetable(home, zone)
	if err != nil {
		return err
	}

	block, err := ctx.getTadoTimetableBlock(home, zone, timetable, ctx.time.Add(5*time.Minute))
	if err != nil {
		return err
	}

	target := block.Setting

	if state.TadoMode == "AWAY" && !block.GeolocationOverride {
		conf, err := ctx.getTadoAwayConfiguration(home, zone)
		if err != nil {
			return err
		}

		if conf.Setting.Power != "OFF" {
			target = conf.Setting
		} else {
			target.Power = "OFF"
		}
	}

	switch target.Mode {
	case "COOL":
		return ctx.smartCool(home, zone, state, target)
	case "DRY":
		return ctx.smartDry(home, zone, state, target)
	case "HEAT":
		return ctx.smartHeat(home, zone, state, target)
	}

	fmt.Printf("OK: (power=%s, mode=%s)\n", target.Power, target.Mode)
	return nil
}

func (ctx *TadoContext) smartCool(home TadoHome, zone TadoZone, state TadoZoneState, target TadoSetting) error {
	curMode := state.Setting.Mode
	curPower := state.Setting.Power
	curTemp := state.SensorDataPoints.InsideTemperature.Celsius
	tgtPower := target.Power
	tgtFan := target.FanSpeed
	tgtTemp := target.Temperature.Celsius

	if curTemp < tgtTemp+0.5 {
		if curPower == "OFF" {
			fmt.Printf("cooling stay off: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOffOverlay(10*time.Minute))
			return err
		}
		if curTemp < tgtTemp-1 {
			fmt.Printf("cooling turn off: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOffOverlay(15*time.Minute))
			return err
		}
	}

	if curMode == "COOL" && tgtPower == "ON" && tgtFan == "AUTO" {
		if curTemp > tgtTemp+4 {
			target.FanSpeed = "HIGH"
			fmt.Printf("cooling boost high: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOverlay(target, 10*time.Minute))
			return err
		}
		if curTemp > tgtTemp+2 {
			target.FanSpeed = "MIDDLE"
			fmt.Printf("cooling boost: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOverlay(target, 10*time.Minute))
			return err
		}
	}

	fmt.Printf("cooling OK: (tgt=%v°C, cur=%v°C, fan=%s, mode=%s)\n", tgtTemp, curTemp, tgtFan, curMode)
	return nil
}

func (ctx *TadoContext) smartHeat(home TadoHome, zone TadoZone, state TadoZoneState, target TadoSetting) error {
	curMode := state.Setting.Mode
	curPower := state.Setting.Power
	curTemp := state.SensorDataPoints.InsideTemperature.Celsius
	tgtPower := target.Power
	tgtFan := target.FanSpeed
	tgtTemp := target.Temperature.Celsius

	if curTemp > tgtTemp-0.5 {
		if curPower == "OFF" {
			fmt.Printf("heating stay off: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOffOverlay(10*time.Minute))
			return err
		}
		if curTemp > tgtTemp+1 {
			fmt.Printf("heating turn off: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOffOverlay(15*time.Minute))
			return err
		}
	}

	if curMode == "HEAT" && tgtPower == "ON" && tgtFan == "AUTO" {
		if curTemp < tgtTemp-4 {
			target.FanSpeed = "HIGH"
			fmt.Printf("heating boost high: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOverlay(target, 10*time.Minute))
			return err
		}
		if curTemp < tgtTemp-2 {
			target.FanSpeed = "MIDDLE"
			fmt.Printf("heating boost: (tgt=%v°C, cur=%v°C)\n", tgtTemp, curTemp)
			_, err := ctx.putTadoOverlay(home, zone, makeOverlay(target, 10*time.Minute))
			return err
		}
	}

	fmt.Printf("heating OK: (tgt=%v°C, cur=%v°C, fan=%s, mode=%s)\n", tgtTemp, curTemp, tgtFan, curMode)
	return nil
}

func (ctx *TadoContext) smartDry(home TadoHome, zone TadoZone, state TadoZoneState, target TadoSetting) error {
	curPower := state.Setting.Power
	curRH := state.SensorDataPoints.Humidity.Percentage

	if 0 < curRH && curRH < 50 {
		if curPower == "OFF" {
			fmt.Printf("drying stay off: (rh=%v%%)\n", curRH)
			_, err := ctx.putTadoOverlay(home, zone, makeOffOverlay(10*time.Minute))
			return err
		}
		if curRH < 40 {
			fmt.Printf("drying turn off: (rh=%v%%)\n", curRH)
			_, err := ctx.putTadoOverlay(home, zone, makeOffOverlay(15*time.Minute))
			return err
		}
	}

	fmt.Printf("drying OK: (rh=%v%%)\n", curRH)
	return nil
}

func makeContext() TadoContext {
	file, err := os.Open("config.json")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	var config TadoConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		panic(fmt.Sprintf("Decode config.json: %s", err))
	}

	res, err := http.PostForm("https://auth.tado.com/oauth/token", url.Values{
		"client_id":     {"public-api-preview"},
		"client_secret": {"4HJGRffVR8xb3XdEUQpjgZ1VplJi6Xgw"},
		"username":      {config.Username},
		"password":      {config.Password},
		"grant_type":    {"password"},
		"scope":         {"home.user"},
	})

	if err != nil {
		panic(fmt.Sprintf("POST /oauth/token: %v", err))
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("POST /oauth/token: %s", http.StatusText(res.StatusCode)))
	}

	var token *oauth2.Token
	if err := json.NewDecoder(res.Body).Decode(&token); err != nil {
		panic(fmt.Sprintf("Decode /oauth/token: %s", err))
	}

	conf := &oauth2.Config{
		ClientID:     "public-api-preview",
		ClientSecret: "4HJGRffVR8xb3XdEUQpjgZ1VplJi6Xgw",
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://auth.tado.com/oauth/token",
		},
	}

	return TadoContext{
		tado: conf.Client(context.Background(), token),
		time: time.Now(),
	}
}

func makeOffOverlay(duration time.Duration) TadoOverlay {
	return makeOverlay(TadoSetting{Type: "AIR_CONDITIONING", Power: "OFF"}, duration)
}

func makeOverlay(setting TadoSetting, duration time.Duration) TadoOverlay {
	overlay := TadoOverlay{
		Setting:     setting,
		Termination: TadoTermination{Type: "TIMER"},
	}
	overlay.Termination.DurationInSeconds = int(duration / time.Second)
	return overlay
}
