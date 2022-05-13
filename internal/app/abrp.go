package app

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var abrpSendQuit = make(chan bool)

func abrpSendActivate(car *Car, endTime time.Time) {
	car.abrpSendActive = true
	car.abrpUpdatesEndTime = endTime

	go abrpSendLoop(car, endTime)
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

			if !endTime.IsZero() && time.Now().After(endTime) {
				log.Println("Stop sending to ABRP (reason: time's up)...")
				abrpSendLoopStop(car)
				return
			}

			if timer%2 != 0 {
				continue
			}

			if car.state != car.previousState {
				timer = 600
			}

			car.abrpData["utc"] = int(time.Now().UTC().Unix())

			if car.state == "parked" || car.state == "online" || car.state == "suspended" || car.state == "asleep" {
				delete(car.abrpData, "kwh_charged")
				if timer%600 == 0 || timer > 600 {
					// parked, update every 10 minutes
					abrpSend(car)
					timer = 0
				}
			} else if car.state == "charging" {
				if timer%10 == 0 {
					// charging, update every 10 seconds
					abrpSend(car)
				}
			} else if car.state == "driving" {
				if timer%2 == 0 {
					// driving, update every 2 seconds
					abrpSend(car)
				}
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

	data, error := json.Marshal(map[string]interface{}{"tlm": car.abrpData})
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
