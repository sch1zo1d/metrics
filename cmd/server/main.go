package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type MemStorage struct {
	Data struct{
		gauge GaugeMetric
		counter CounterMetric
	}
}

type gauge float64
type counter int64

type GaugeMetric map[string]gauge
type CounterMetric map[string]counter

type Storage interface {
	Add()
	Delete()
}

var (
	mu sync.Mutex
	db = MemStorage{
    Data: struct {
        gauge   GaugeMetric
        counter CounterMetric
    }{
        gauge:   make(GaugeMetric),
        counter: make(CounterMetric),
    },
}

)

func HandlerMetric(w http.ResponseWriter, req *http.Request){
	if req.Method != http.MethodPost || req.Header.Get("content-type") != "text/plain"{
		w.WriteHeader(http.StatusMethodNotAllowed)
        return
	}
	vars := mux.Vars(req)
	
	mu.Lock()
	badReq := 0
	if vars["name"] == "" {w.WriteHeader(http.StatusNotFound); return}
	switch vars["type"] {
	case "counter": 
		if val, err := strconv.ParseInt(vars["value"], 10, 64); err==nil {
			db.Data.counter[vars["name"]] += counter(val)
		} else {badReq = 1}
	case "gauge": 
		if val, err := strconv.ParseFloat(vars["value"], 64); err==nil {
			db.Data.gauge[vars["name"]] = gauge(val)
		} else {badReq = 1}
	default: badReq = 1
	}
	if badReq == 1 {w.WriteHeader(http.StatusBadRequest); return}
	mu.Unlock()
	w.Header().Set("content-type", "text/plain")
	w.Header().Set("charset", "utf-8")
	w.WriteHeader(http.StatusOK)
}

func main() {
	// init
	mux := mux.NewRouter()
	mux.HandleFunc("/update/{type}/{name}/{value}", HandlerMetric)
    log.Fatal(http.ListenAndServe(":8080", mux))
}
