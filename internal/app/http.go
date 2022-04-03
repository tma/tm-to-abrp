package app

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

var indexTemplate = template.Must(template.ParseFiles("web/templates/index.html"))

func indexHandler(car Car) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

func abrpSendContinuousHandler(car Car) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if car.abrpSendActive {
			abrpSendDeactivate()
		} else {
			endTimeString := r.FormValue("endtime")

			var endTime time.Time
			if endTimeString != "" {
				location, _ := time.LoadLocation(os.Getenv("TZ"))
				endTime, _ = time.ParseInLocation("2006-01-02T15:04", endTimeString, location)
			}

			abrpSendActivate(car, endTime)
		}

		http.Redirect(w, r, rootPath, http.StatusFound)
	}
}

func abrpSendNowHandler(car Car) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		abrpSend(car)

		http.Redirect(w, r, rootPath, http.StatusFound)
	}
}

var rootPath string

func WebStart(port string, car Car) {
	log.Println("Starting HTTP server on port " + port)

	router := mux.NewRouter()

	fs := http.FileServer(http.Dir("web/public"))
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
	router.HandleFunc("", indexHandler(car))
	router.HandleFunc("/", indexHandler(car))
	router.HandleFunc("/abrp-send-continuous", abrpSendContinuousHandler(car)).Methods("POST")
	router.HandleFunc("/abrp-send-now", abrpSendNowHandler(car)).Methods("POST")

	http.ListenAndServe(":"+port, router)
}
