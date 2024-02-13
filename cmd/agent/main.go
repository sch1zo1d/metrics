package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverAddress  = "http://localhost:8080"
)

type gauge float64
type counter int64

type GaugeMetric map[string]gauge
type CounterMetric map[string]counter

var (
	randSrc     = rand.NewSource(time.Now().UnixNano())
	rnd         = rand.New(randSrc)
	randomValue = gauge(rnd.Float64())

	wg     sync.WaitGroup
	stopCh = make(chan struct{})

	db = struct {
		gauge   GaugeMetric
		counter CounterMetric
	}{
		gauge:   make(GaugeMetric),
		counter: make(CounterMetric),
	}
)

func gatherMetrics() {
	defer wg.Done()
	mem := runtime.MemStats{}
	for {
		select {
		case <-stopCh:
			return
		default:
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
			db.gauge["RandomValue"] = randomValue
			updateRandomValue()
			time.Sleep(pollInterval)
		}
	}
}

func sendMetrics() {
	for {
		select {
		case <-stopCh:
			return
		default:
			t := reflect.ValueOf(db)

			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)

				if field.Kind() == reflect.Map {
					iter := field.MapRange()
					for iter.Next() {
						metricName := iter.Key().String()
						metricValue := iter.Value()

						url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddress, t.Type().Field(i).Name, metricName, metricValue)
						resp, err := http.Post(url, "text/plain", http.NoBody)
						if err != nil {
							fmt.Printf("Ошибка при отправке метрики: %s\n", err.Error())
							continue
						}
						defer resp.Body.Close()
						if resp.StatusCode != http.StatusOK {
							fmt.Printf("Ошибка при отправке метрики: %s\n", resp.Status)
						}
						fmt.Println(resp)
					}
				}
			}
			time.Sleep(reportInterval)
		}
	}
}

func updateRandomValue() {
	randomValue = gauge(rnd.Float64())
}

func main() {

	wg.Add(2)

	go gatherMetrics()
	go sendMetrics()

	wg.Wait()

}
