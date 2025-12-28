package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()

	// /temperature?location=...&sensorId=...
	mux.HandleFunc("/temperature", temperatureQueryHandler)

	// /temperature/{id}
	mux.HandleFunc("/temperature/", temperatureByIDHandler)

	addr := ":8081"
	log.Printf("temperature-api started on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func temperatureQueryHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	location := q.Get("location")

	// поддержим оба варианта, чтобы не ловить мелкие расхождения
	sensorID := q.Get("sensorId")
	if sensorID == "" {
		sensorID = q.Get("sensor_id")
	}

	location, sensorID = normalizeLocationAndSensor(location, sensorID)

	resp := buildResponse(location, sensorID)
	writeJSON(w, http.StatusOK, resp)
}

func temperatureByIDHandler(w http.ResponseWriter, r *http.Request) {
	// ожидаем путь ровно вида: /temperature/{id}
	idPart := strings.TrimPrefix(r.URL.Path, "/temperature/")
	idPart = strings.Trim(idPart, "/")

	// если пришли просто на /temperature/ (без id) — отдадим 404, чтобы не путать с /temperature
	if idPart == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "missing sensor id in path, use /temperature/{id}",
		})
		return
	}

	// на всякий: берём только первый сегмент
	if i := strings.IndexByte(idPart, '/'); i >= 0 {
		idPart = idPart[:i]
	}

	sensorID := idPart

	// location можно опционально передать query-параметром, иначе выведем по sensorID
	location := r.URL.Query().Get("location")

	location, sensorID = normalizeLocationAndSensor(location, sensorID)

	resp := buildResponse(location, sensorID)
	writeJSON(w, http.StatusOK, resp)
}

func normalizeLocationAndSensor(location, sensorID string) (string, string) {
	// If no location is provided, use a default based on sensor ID
	if location == "" {
		switch sensorID {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	// If no sensor ID is provided, generate one based on location
	if sensorID == "" {
		switch location {
		case "Living Room":
			sensorID = "1"
		case "Bedroom":
			sensorID = "2"
		case "Kitchen":
			sensorID = "3"
		default:
			sensorID = "0"
		}
	}

	return location, sensorID
}

func buildResponse(location, sensorID string) TemperatureResponse {
	// Рандомная температура: 15.0..30.0 с шагом 0.1
	v := 15.0 + rand.Float64()*15.0
	v = float64(int(v*10)) / 10

	status := "ok"
	if location == "Unknown" || sensorID == "0" {
		status = "unknown_sensor"
	}

	return TemperatureResponse{
		Value:       v,
		Unit:        "C",
		Timestamp:   time.Now().UTC(),
		Location:    location,
		Status:      status,
		SensorID:    sensorID,
		SensorType:  "temperature",
		Description: "Random temperature reading",
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
