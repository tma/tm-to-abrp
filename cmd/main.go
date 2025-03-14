package main

import (
	"os"
	"strconv"
	"tm-to-abrp/internal/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mqttAddress := os.Getenv("MQTT")
	if mqttAddress == "" {
		mqttAddress = "tcp://localhost:1883"
	}

	// Get MQTT credentials
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")

	// Get TLS settings
	mqttUseTLS := false
	useTLSStr := os.Getenv("MQTT_TLS")
	if useTLSStr != "" {
		var err error
		mqttUseTLS, err = strconv.ParseBool(useTLSStr)
		if err != nil {
			mqttUseTLS = false
		}
	}

	mqttTlsSkipVerify := false
	tlsSkipStr := os.Getenv("MQTT_TLS_SKIP_VERIFY")
	if tlsSkipStr != "" {
		var err error
		mqttTlsSkipVerify, err = strconv.ParseBool(tlsSkipStr)
		if err != nil {
			mqttTlsSkipVerify = false
		}
	}

	if os.Getenv("TZ") == "" {
		panic("TZ environment variable not set")
	}

	if os.Getenv("ABRP_TOKEN") == "" {
		panic("ABRP_TOKEN environment variable not set")
	}

	if os.Getenv("ABRP_API_KEY") == "" {
		panic("ABRP_API_KEY environment variable not set")
	}

	carNumber := os.Getenv("TM_CAR_NUMBER")
	if carNumber == "" {
		carNumber = "1"
	}

	car := *app.NewCar(
		carNumber,
		os.Getenv("ABRP_CAR_MODEL"),
		os.Getenv("ABRP_TOKEN"),
		os.Getenv("ABRP_API_KEY"),
	)

	app.MessagesSubscribe(mqttAddress, mqttUsername, mqttPassword, mqttUseTLS, mqttTlsSkipVerify, &car)
	app.WebStart(port, &car)
}
