package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

	. "github.com/sch1zo1d/metrics/internal/constant"
)

const (
	htp = "http://"
)

var (
	randSrc     = rand.NewSource(time.Now().UnixNano())
	rnd         = rand.New(randSrc)
	randomValue = Gauge(rnd.Float64())

	wg             sync.WaitGroup
	randomValueMux sync.RWMutex
	dbMux          sync.RWMutex

	stopCh = make(chan struct{})

	db = struct {
		gauge   GaugeMetric
		counter CounterMetric
	}{
		gauge:   make(GaugeMetric),
		counter: make(CounterMetric),
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
			db.gauge["Alloc"] = Gauge(mem.Alloc)
			db.gauge["BuckHashSys"] = Gauge(mem.BuckHashSys)
			db.gauge["Frees"] = Gauge(mem.Frees)
			db.gauge["GCCPUFraction"] = Gauge(mem.GCCPUFraction)
			db.gauge["GCSys"] = Gauge(mem.GCSys)
			db.gauge["HeapAlloc"] = Gauge(mem.HeapAlloc)
			db.gauge["HeapIdle"] = Gauge(mem.HeapIdle)
			db.gauge["HeapInuse"] = Gauge(mem.HeapInuse)
			db.gauge["HeapObjects"] = Gauge(mem.HeapObjects)
			db.gauge["HeapReleased"] = Gauge(mem.HeapReleased)
			db.gauge["HeapSys"] = Gauge(mem.HeapSys)
			db.gauge["LastGC"] = Gauge(mem.LastGC)
			db.gauge["Lookups"] = Gauge(mem.Lookups)
			db.gauge["MCacheInuse"] = Gauge(mem.MCacheInuse)
			db.gauge["MCacheSys"] = Gauge(mem.MCacheSys)
			db.gauge["MSpanInuse"] = Gauge(mem.MSpanInuse)
			db.gauge["MSpanSys"] = Gauge(mem.MSpanSys)
			db.gauge["Mallocs"] = Gauge(mem.Mallocs)
			db.gauge["NextGC"] = Gauge(mem.NextGC)
			db.gauge["NumForcedGC"] = Gauge(mem.NumForcedGC)
			db.gauge["NumGC"] = Gauge(mem.NumGC)
			db.gauge["OtherSys"] = Gauge(mem.OtherSys)
			db.gauge["PauseTotalNs"] = Gauge(mem.PauseTotalNs)
			db.gauge["StackInuse"] = Gauge(mem.StackInuse)
			db.gauge["StackSys"] = Gauge(mem.StackSys)
			db.gauge["Sys"] = Gauge(mem.Sys)
			db.gauge["TotalAlloc"] = Gauge(mem.TotalAlloc)

			db.counter["PollCount"] += Counter(1)
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
						// metric := new(Metrics{ssdf, ad, ad})
						metric := &Metrics{ID: iter.Key().String(), MType: t.Type().Field(i).Name}
						switch t.Type().Field(i).Name {
						case CounterS:
							metric.Delta = new(int64)
							*metric.Delta = iter.Value().Int()
						case GaugeS:
							metric.Value = new(float64)
							*metric.Value = iter.Value().Float()
						}
						// log.Println(metric, *metric.Value, *metric.Delta)
						// gzip.
						js, err := json.Marshal(metric)
						if err != nil {
							log.Printf("Ошибка при сериализации метрики: %s\n", err.Error())
							continue
						}
						var buf bytes.Buffer
						gz := gzip.NewWriter(&buf)
						gz.Write(js)
						gz.Close()
						url := fmt.Sprintf("%s%s/update/", htp, serverAddress)
						resp, err := http.Post(url, "application/json", &buf)
						if err != nil {
							log.Printf("Ошибка при отправке метрики: %s\n", err.Error())
							continue
						}
						resp.Header.Set("Content-Encoding", "gzip")
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
	randomValue = Gauge(rnd.Float64())
	randomValueMux.Unlock()
}

func main() {

	wg.Add(2)
	parseFlags()
	go gatherMetrics()
	go sendMetrics()

	wg.Wait()
}
