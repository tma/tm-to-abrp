package app

import (
	"strconv"
	"sync"
	"time"
)

type Car struct {
	number             string
	state              string
	previousState      string
	tmData             *sync.Map
	abrpData           *sync.Map
	abrpSendActive     bool
	abrpUpdatesEndTime time.Time
	abrpToken          string
	abrpApiKey         string
}

func NewCar(number string, carModel string, abrpToken string, abrpApiKey string) *Car {
	return &Car{
		number: number,
		tmData: new(sync.Map),
		abrpData: new(sync.Map),
		// 	"car_model":           carModel,
		// 	"utc":                 0,
		// 	"soc":                 0,
		// 	"power":               0,
		// 	"speed":               0,
		// 	"lat":                 "",
		// 	"lon":                 "",
		// 	"elevation":           "",
		// 	"heading":             "",
		// 	"is_charging":         0,
		// 	"is_dcfc":             0,
		// 	"is_parked":           0,
		// 	"est_battery_range":   "",
		// 	"ideal_battery_range": "",
		// 	"ext_temp":            "",
		// 	"tlm_type":            "api",
		// 	"voltage":             0,
		// 	"current":             0,
		// 	"kwh_charged":         0,
		// 	"odometer":            "",
		// },
		abrpSendActive: false,
		abrpToken:      abrpToken,
		abrpApiKey:     abrpApiKey,
	}
}

func updateCarTmData(car *Car, topic string, payload string) {
	car.tmData.Store(topic, payload)
}

func updateCarAbrpData(car *Car, topic string, payload string) {
	switch topic {
	case "latitude":
		car.abrpData.Store("lat", payload)
	case "longitude":
		car.abrpData.Store("lon", payload)
	case "elevation":
		car.abrpData.Store("elevation", payload)
	case "heading":
		car.abrpData.Store("heading", payload)
	case "speed":
		value, _ := strconv.Atoi(payload)
		car.abrpData.Store("speed", value)
	case "outside_temp":
		car.abrpData.Store("ext_temp", payload)
	case "odometer":
		car.abrpData.Store("odometer", payload)
	case "ideal_battery_range_km":
		car.abrpData.Store("ideal_battery_range", payload)
	case "est_battery_range_km":
		car.abrpData.Store("est_battery_range", payload)
	case "usable_battery_level":
		car.abrpData.Store("soc", payload)
	case "charge_energy_added":
		car.abrpData.Store("kwh_charged", payload)
	case "power":
		power, _ := strconv.Atoi(payload)
		car.abrpData.Store("power", power)

		isCharging, _ := car.abrpData.Load("is_charging")
		if (isCharging == 1) && (power < -22) {
			car.abrpData.Store("is_dcfc", 1)
		}
	case "charger_power":
		if payload != "" && payload != "0" {
			car.abrpData.Store("is_charging", 1)

			chargerPower, _ := strconv.Atoi(payload)
			if chargerPower > 22 {
				car.abrpData.Store("is_dcfc", 1)
			}
		}
	case "charger_actual_current":
		if payload != "" {
			current, _ := strconv.Atoi(payload)
			if current > 0 {
				car.abrpData.Store("current", payload)
			} else {
				car.abrpData.Delete("current")
			}
		}
	case "charger_voltage":
		if payload != "" {
			voltage, _ := strconv.Atoi(payload)
			if voltage > 4 {
				car.abrpData.Store("voltage", payload)
			} else {
				car.abrpData.Delete("voltage")
			}
		}
	case "state":
		car.state = payload
		if car.state == "driving" {
			car.abrpData.Store("is_parked", 0)
			car.abrpData.Store("is_charging", 0)
			car.abrpData.Store("is_dcfc", 0)
		} else if car.state == "charging" {
			car.abrpData.Store("is_parked", 1)
			car.abrpData.Store("is_charging", 1)
			car.abrpData.Store("is_dcfc", 0)
		} else if car.state == "supercharging" {
			car.abrpData.Store("is_parked", 1)
			car.abrpData.Store("is_charging", 1)
			car.abrpData.Store("is_dcfc", 1)
		} else if car.state == "online" || car.state == "suspended" || car.state == "asleep" {
			car.abrpData.Store("is_parked", 1)
			car.abrpData.Store("is_charging", 0)
			car.abrpData.Store("is_dcfc", 0)
		}
	case "shift_state":
		if payload == "P" {
			car.abrpData.Store("is_parked", 1)
		} else if payload == "D" || payload == "R" {
			car.abrpData.Store("is_parked", 0)
		}
	}
}
