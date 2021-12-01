package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
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

var car Car = Car{
	tmData: map[string]interface{}{},
	abrpData: map[string]interface{}{
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
		"battery_range":       "",
		"ideal_battery_range": "",
		"ext_temp":            "",
		"tlm_type":            "api",
		"voltage":             0,
		"current":             0,
		"kwh_charged":         0,
		"odometer":            "",
	},
}

func mqttSubscribe(mqttAddress string) {
	log.Println("Connecting to MQTT server on " + mqttAddress)

	opts := mqtt.NewClientOptions().AddBroker(mqttAddress).SetClientID("tm-to-abrp-car" + car.number)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := client.Subscribe("teslamate/cars/"+car.number+"/#", 0, mqttCarMessage); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("Listening for updates of car number " + car.number)
}

var mqttCarMessage mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	topic := strings.Split(msg.Topic(), "/")[3]
	payload := string(msg.Payload())

	updateCarTmData(topic, payload)
	updateCarAbrpData(topic, payload)
}

func updateCarTmData(topic string, payload string) {
	car.tmData[topic] = payload
}

func updateCarAbrpData(topic string, payload string) {
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

var indexTemplate = template.Must(template.ParseFiles("templates/index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	carTmData, _ := json.MarshalIndent(car.tmData, "", "  ")
	carAbrpData, _ := json.MarshalIndent(car.abrpData, "", "  ")

	var carAbrpSendContinuousButtonText string
	if car.abrpSendActive {
		carAbrpSendContinuousButtonText = "Stop"
	} else {
		carAbrpSendContinuousButtonText = "Start"
	}

	var carAbrpUpdatesEndTimeString string
	if car.abrpUpdatesEndTime.IsZero() {
		carAbrpUpdatesEndTimeString = ""
	} else {
		carAbrpUpdatesEndTimeString = car.abrpUpdatesEndTime.Format("2006-01-02T15:04")
	}

	indexTemplate.Execute(w, map[string]interface{}{
		"rootPath":                        rootPath,
		"carState":                        car.state,
		"carPreviousState":                car.previousState,
		"carTmData":                       string(carTmData),
		"carAbrpData":                     string(carAbrpData),
		"carAbrpUpdatesEndTimeString":     carAbrpUpdatesEndTimeString,
		"carAbrpSendContinuousButtonText": carAbrpSendContinuousButtonText,
	})
}

func abrpSendContinuousHandler(w http.ResponseWriter, r *http.Request) {
	if car.abrpSendActive {
		abrpSendDeactivate()
	} else {
		endTimeString := r.FormValue("endtime")

		var endTime time.Time
		if endTimeString != "" {
			location, _ := time.LoadLocation(os.Getenv("TZ"))
			endTime, _ = time.ParseInLocation("2006-01-02T15:04", endTimeString, location)
		}

		abrpSendActivate(endTime)
	}

	http.Redirect(w, r, rootPath, http.StatusFound)
}

func abrpSendNowHandler(w http.ResponseWriter, r *http.Request) {
	abrpSend()

	http.Redirect(w, r, rootPath, http.StatusFound)
}

var rootPath string

func httpStart(port string) {
	log.Println("Starting HTTP server on port " + port)

	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("public"))
	publicHandler := http.StripPrefix("/public/", fs)

	pathPrefix := os.Getenv("PATH_PREFIX")
	if pathPrefix != "" {
		rootPath = pathPrefix
		router = router.PathPrefix(pathPrefix).Subrouter()
		publicHandler = http.StripPrefix(pathPrefix+"/public/", fs)
	} else {
		rootPath = ""
	}

	router.PathPrefix("/public/").Handler(publicHandler)
	router.HandleFunc("", indexHandler)
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/abrp-send-continuous", abrpSendContinuousHandler).Methods("POST")
	router.HandleFunc("/abrp-send-now", abrpSendNowHandler).Methods("POST")

	http.ListenAndServe(":"+port, router)
}

var abrpSendQuit = make(chan bool)

func abrpSendActivate(endTime time.Time) {
	car.abrpSendActive = true
	car.abrpUpdatesEndTime = endTime

	go abrpSendLoop(endTime)
}

func abrpSendDeactivate() {
	abrpSendQuit <- true
}

const abrpSendInterval = 1 * time.Second

func abrpSendLoop(endTime time.Time) {
	log.Println("Start sending to ABRP...")

	timer := -1

	for {
		select {
		case <-abrpSendQuit:
			log.Println("Stop sending to ABRP (reason: manual)...")
			abrpSendLoopStop()
			return
		default:
			timer++
			time.Sleep(abrpSendInterval)

			if !endTime.IsZero() && time.Now().After(endTime) {
				log.Println("Stop sending to ABRP (reason: time's up)...")
				abrpSendLoopStop()
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
					abrpSend()
					timer = 0
				}
			} else if car.state == "charging" {
				if timer%10 == 0 {
					// charging, update every 10 seconds
					abrpSend()
				}
			} else if car.state == "driving" {
				if timer%2 == 0 {
					// driving, update every 2 seconds
					abrpSend()
				}
			}

			car.previousState = car.state
		}
	}
}

func abrpSendLoopStop() {
	car.abrpSendActive = false
	car.abrpUpdatesEndTime = time.Time{}
}

const abrpUrl = "https://api.iternio.com/1/tlm/send"

func abrpSend() {
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

	car.number = carNumber
	car.abrpData["car_model"] = os.Getenv("ABRP_CAR_MODEL")
	car.abrpToken = os.Getenv("ABRP_TOKEN")
	car.abrpApiKey = os.Getenv("ABRP_API_KEY")
	car.abrpSendActive = false

	mqttSubscribe(mqttAddress)
	httpStart(port)
}
