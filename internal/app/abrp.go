package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var abrpSendQuit = make(chan bool)

func areWeSendingIndefinitely() bool {
	return os.Getenv("ABRP_SEND_INDEFINITELY") == "1"
}

func abrpSendActivate(car *Car, endTime time.Time) {
	car.abrpSendActive = true
	car.abrpUpdatesEndTime = endTime

	if areWeSendingIndefinitely() {
		go abrpSendLoop(car, time.Time{})
	} else {
		go abrpSendLoop(car, endTime)
	}
}

func abrpSendDeactivate() {
	abrpSendQuit <- true
}

const abrpSendInterval = 1 * time.Second

func abrpSendLoop(car *Car, endTime time.Time) {
	log.Println("Start sending to ABRP...")

	timer := -1

	for {
		select {
		case <-abrpSendQuit:
			log.Println("Stop sending to ABRP (reason: manual)...")
			abrpSendLoopStop(car)
			return
		default:
			timer++
			time.Sleep(abrpSendInterval)

			if !areWeSendingIndefinitely() && !endTime.IsZero() && time.Now().After(endTime) {
				log.Println("Stop sending to ABRP (reason: time's up)...")
				abrpSendLoopStop(car)
				return
			}

			if car.state != car.previousState {
				timer = 30
			}

			car.abrpData.Store("utc", int(time.Now().UTC().Unix()))

			if car.state == "parked" || car.state == "online" || car.state == "suspended" || car.state == "asleep" {
				car.abrpData.Delete("kwh_charged")
				if timer%30 == 0 || timer > 30 {
					// parked, update every 30 seconds (avoid being offline for ABRP)
					abrpSend(car)
					timer = 0
				}
			} else if car.state == "charging" {
				if timer%4 == 0 {
					// charging, update every 4 seconds
					abrpSend(car)
				}
			} else if car.state == "driving" {
				// driving, update every second
				abrpSend(car)
			}

			car.previousState = car.state
		}
	}
}

func abrpSendLoopStop(car *Car) {
	car.abrpSendActive = false
	car.abrpUpdatesEndTime = time.Time{}
}

const abrpUrl = "https://api.iternio.com/1/tlm/send"

func abrpSend(car *Car) {
	log.Println("Sending to ABRP...")

	m := map[string]interface{}{}
	car.abrpData.Range(func(key, value interface{}) bool {
		m[fmt.Sprint(key)] = value
		return true
	})

	data, error := json.Marshal(map[string]interface{}{"tlm": m})
	if error != nil {
		log.Println("Error marshalling data: " + error.Error())
		return
	}

	request, error := http.NewRequest("POST", abrpUrl+"?token="+car.abrpToken, bytes.NewBuffer(data))
	if error != nil {
		log.Println("Error creating request: " + error.Error())
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "APIKEY "+car.abrpApiKey)

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		log.Println("Error sending request: " + error.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Println("Error sending request: " + response.Status)
		return
	}
}
