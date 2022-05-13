package main

import (
	"os"
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

	app.MessagesSubscribe(mqttAddress, &car)
	app.WebStart(port, &car)
}
