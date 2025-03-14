package app

import (
	"crypto/tls"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func MessagesSubscribe(mqttAddress string, username string, password string, useTLS bool, tlsSkipVerify bool, car *Car) {
	log.Println("Connecting to MQTT server on " + mqttAddress)

	opts := mqtt.NewClientOptions().AddBroker(mqttAddress).SetClientID("tm-to-abrp-car" + car.number)

	// Add username and password if provided
	if username != "" {
		opts.SetUsername(username)
		if password != "" {
			opts.SetPassword(password)
		}
	}

	// Configure TLS if needed
	if useTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: tlsSkipVerify,
		}
		opts.SetTLSConfig(tlsConfig)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := client.Subscribe("teslamate/cars/"+car.number+"/#", 0, func(c mqtt.Client, m mqtt.Message) {
		topic := strings.Split(m.Topic(), "/")[3]
		payload := string(m.Payload())

		updateCarTmData(car, topic, payload)
		updateCarAbrpData(car, topic, payload)
	}); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("Listening for updates of car number " + car.number)
}
