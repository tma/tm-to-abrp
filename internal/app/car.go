package app

import (
	"strconv"
	"time"
)

type Car struct {
	number             string
	state              string
	previousState      string
	tmData             map[string]interface{}
	abrpData           map[string]interface{}
	abrpSendActive     bool
	abrpUpdatesEndTime time.Time
	abrpToken          string
	abrpApiKey         string
}

func NewCar(number string, carModel string, abrpToken string, abrpApiKey string) *Car {
	return &Car{
		number: number,
		tmData: map[string]interface{}{},
		abrpData: map[string]interface{}{
			"car_model":           carModel,
			"utc":                 0,
			"soc":                 0,
			"power":               0,
			"speed":               0,
			"lat":                 "",
			"lon":                 "",
			"elevation":           "",
			"heading":             "",
			"is_charging":         0,
			"is_dcfc":             0,
			"is_parked":           0,
			"est_battery_range":   "",
			"ideal_battery_range": "",
			"ext_temp":            "",
			"tlm_type":            "api",
			"voltage":             0,
			"current":             0,
			"kwh_charged":         0,
			"odometer":            "",
		},
		abrpSendActive: false,
		abrpToken:      abrpToken,
		abrpApiKey:     abrpApiKey,
	}
}

func updateCarTmData(car *Car, topic string, payload string) {
	car.tmData[topic] = payload
}

func updateCarAbrpData(car *Car, topic string, payload string) {
	switch topic {
	case "latitude":
		car.abrpData["lat"] = payload
	case "longitude":
		car.abrpData["lon"] = payload
	case "elevation":
		car.abrpData["elevation"] = payload
	case "heading":
		car.abrpData["heading"] = payload
	case "speed":
		car.abrpData["speed"], _ = strconv.Atoi(payload)
	case "outside_temp":
		car.abrpData["ext_temp"] = payload
	case "odometer":
		car.abrpData["odometer"] = payload
	case "ideal_battery_range_km":
		car.abrpData["ideal_battery_range"] = payload
	case "est_battery_range_km":
		car.abrpData["battery_range"] = payload
	case "battery_level":
		car.abrpData["soc"] = payload
	case "charge_energy_added":
		car.abrpData["kwh_charged"] = payload
	case "power":
		power, _ := strconv.Atoi(payload)
		car.abrpData["power"] = power

		if (car.abrpData["is_charging"] == 1) && (power < -22) {
			car.abrpData["is_dcfc"] = 1
		}
	case "charger_power":
		if payload != "" && payload != "0" {
			car.abrpData["is_charging"] = 1

			chargerPower, _ := strconv.Atoi(payload)
			if chargerPower > 22 {
				car.abrpData["is_dcfc"] = 1
			}
		}
	case "charger_actual_current":
		if payload != "" {
			current, _ := strconv.Atoi(payload)
			if current > 0 {
				car.abrpData["current"] = payload
			} else {
				delete(car.abrpData, "current")
			}
		}
	case "charger_voltage":
		if payload != "" {
			voltage, _ := strconv.Atoi(payload)
			if voltage > 4 {
				car.abrpData["voltage"] = payload
			} else {
				delete(car.abrpData, "voltage")
			}
		}
	case "state":
		car.state = payload
		if car.state == "driving" {
			car.abrpData["is_parked"] = 0
			car.abrpData["is_charging"] = 0
			car.abrpData["is_dcfc"] = 0
		} else if car.state == "charging" {
			car.abrpData["is_parked"] = 1
			car.abrpData["is_charging"] = 1
			car.abrpData["is_dcfc"] = 0
		} else if car.state == "supercharging" {
			car.abrpData["is_parked"] = 1
			car.abrpData["is_charging"] = 1
			car.abrpData["is_dcfc"] = 1
		} else if car.state == "online" || car.state == "suspended" || car.state == "asleep" {
			car.abrpData["is_parked"] = 1
			car.abrpData["is_charging"] = 0
			car.abrpData["is_dcfc"] = 0
		}
	case "shift_state":
		if payload == "P" {
			car.abrpData["is_parked"] = 1
		} else if payload == "D" || payload == "R" {
			car.abrpData["is_parked"] = 0
		}
	}
}
