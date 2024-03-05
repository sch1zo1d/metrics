package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	htp = "http://"
)

type gauge float64
type counter int64

type gaugeMetric map[string]gauge
type counterMetric map[string]counter

var (
	randSrc     = rand.NewSource(time.Now().UnixNano())
	rnd         = rand.New(randSrc)
	randomValue = gauge(rnd.Float64())

	wg     sync.WaitGroup
	randomValueMux sync.RWMutex
	dbMux sync.RWMutex

	stopCh = make(chan struct{})

	db = struct {
		gauge   gaugeMetric
		counter counterMetric
	}{
		gauge:   make(gaugeMetric),
		counter: make(counterMetric),
	}
	pollInterval   int
	reportInterval int
	serverAddress  string
)

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 10, "The frequency of sending metrics to the server (default is 10 seconds)")
	flag.IntVar(&pollInterval, "p", 2, "The polling frequency of metrics from the runtime package (default is 2 seconds)")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
        serverAddress = envRunAddr
    }
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
        reportInterval, _ = strconv.Atoi(envReportInterval)
    }
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
        pollInterval, _ = strconv.Atoi(envPollInterval)
    }

}

func gatherMetrics() {
	defer wg.Done()
	mem := runtime.MemStats{}
	for {
		select {
		case <-stopCh:
			return
		default:
			dbMux.Lock()
			db.gauge["Alloc"] = gauge(mem.Alloc)
			db.gauge["BuckHashSys"] = gauge(mem.BuckHashSys)
			db.gauge["Frees"] = gauge(mem.Frees)
			db.gauge["GCCPUFraction"] = gauge(mem.GCCPUFraction)
			db.gauge["GCSys"] = gauge(mem.GCSys)
			db.gauge["HeapAlloc"] = gauge(mem.HeapAlloc)
			db.gauge["HeapIdle"] = gauge(mem.HeapIdle)
			db.gauge["HeapInuse"] = gauge(mem.HeapInuse)
			db.gauge["HeapObjects"] = gauge(mem.HeapObjects)
			db.gauge["HeapReleased"] = gauge(mem.HeapReleased)
			db.gauge["HeapSys"] = gauge(mem.HeapSys)
			db.gauge["LastGC"] = gauge(mem.LastGC)
			db.gauge["Lookups"] = gauge(mem.Lookups)
			db.gauge["MCacheInuse"] = gauge(mem.MCacheInuse)
			db.gauge["MCacheSys"] = gauge(mem.MCacheSys)
			db.gauge["MSpanInuse"] = gauge(mem.MSpanInuse)
			db.gauge["MSpanSys"] = gauge(mem.MSpanSys)
			db.gauge["Mallocs"] = gauge(mem.Mallocs)
			db.gauge["NextGC"] = gauge(mem.NextGC)
			db.gauge["NumForcedGC"] = gauge(mem.NumForcedGC)
			db.gauge["NumGC"] = gauge(mem.NumGC)
			db.gauge["OtherSys"] = gauge(mem.OtherSys)
			db.gauge["PauseTotalNs"] = gauge(mem.PauseTotalNs)
			db.gauge["StackInuse"] = gauge(mem.StackInuse)
			db.gauge["StackSys"] = gauge(mem.StackSys)
			db.gauge["Sys"] = gauge(mem.Sys)
			db.gauge["TotalAlloc"] = gauge(mem.TotalAlloc)

			db.counter["PollCount"] += counter(1)
			randomValueMux.RLock()
			db.gauge["RandomValue"] = randomValue
			randomValueMux.RUnlock()
			dbMux.Unlock()
			updateRandomValue()
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}
}

func sendMetrics() {

	for {
		select {
		case <-stopCh:
			return
		default:
			dbMux.RLock()
			t := reflect.ValueOf(db)

			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)

				if field.Kind() == reflect.Map {
					iter := field.MapRange()
					for iter.Next() {

						url := fmt.Sprintf("%s%s/update/%s/%s/%v", htp, serverAddress, t.Type().Field(i).Name, iter.Key().String(), iter.Value())
						resp, err := http.Post(url, "text/plain", http.NoBody)
						if err != nil {
							log.Printf("Ошибка при отправке метрики: %s\n", err.Error())
							continue
						}
						defer resp.Body.Close()
						if resp.StatusCode != http.StatusOK {
							log.Printf("Ошибка при отправке метрики: %s\n", resp.Status)
						}
						log.Println(resp)
					}
				}
			}
			dbMux.RUnlock()
			time.Sleep(time.Duration(reportInterval) * time.Second)
		}
	}
}

func updateRandomValue() {
	randomValueMux.Lock()
	randomValue = gauge(rnd.Float64())
	randomValueMux.Unlock()
}

func main() {

	wg.Add(2)
	parseFlags()
	go gatherMetrics()
	go sendMetrics()

	wg.Wait()

}
