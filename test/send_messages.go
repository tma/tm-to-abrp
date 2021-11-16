package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messages = map[string]string{
	"latitude":               "8.75",
	"longitude":              "47.65",
	"elevation":              "275",
	"heading":                "92",
	"speed":                  "25",
	"outside_temp":           "23",
	"odometer":               "42",
	"ideal_battery_range_km": "425",
	"est_battery_range_km":   "420",
	"battery_level":          "82",
	"charge_energy_added":    "2.55",
	"power":                  "-52",
	"charger_power":          "52",
	"charger_actual_current": "80",
	"charger_voltage":        "380",
	"state":                  "driving",
	"shift_state":            "P",
}

func main() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("test")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for topic, payload := range messages {
		token := client.Publish("teslamate/cars/1/"+topic, 0, false, payload)
		token.Wait()
	}
}
