package main

import (
	"flag"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type MemStorage struct {
	Data struct {
		gauge   GaugeMetric
		counter CounterMetric
	}
	mu sync.Mutex
}

type gauge float64
type counter int64

type GaugeMetric map[string]gauge
type CounterMetric map[string]counter

type Storage interface {
	AddCounterMetric(name string, value int64)
	AddGaugeMetric(name string, value float64)
	GetMetrics() (CounterMetric, GaugeMetric)
}

var (

	db Storage = &MemStorage{
		Data: struct {
			gauge   GaugeMetric
			counter CounterMetric
		}{
			gauge:   make(GaugeMetric),
			counter: make(CounterMetric),
		},
	}
)

var flagRunAddr string
func parseFlags() {
    flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
    flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
        flagRunAddr = envRunAddr
    }
} 

func (db *MemStorage) AddCounterMetric(name string, value int64) {
	db.mu.Lock()
	db.Data.counter[name] += counter(value)
	db.mu.Unlock()
}

func (db *MemStorage) AddGaugeMetric(name string, value float64) {
	db.mu.Lock()
	db.Data.gauge[name] = gauge(value)
	db.mu.Unlock()
}

func (db *MemStorage) GetMetrics() (CounterMetric, GaugeMetric) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.Data.counter, db.Data.gauge
}


func HandlerListMetrics(c *gin.Context) {
	counterMetrics, gaugeMetrics := db.GetMetrics()

	data := struct {
		CounterMetrics map[string]counter
		GaugeMetrics   map[string]gauge
	}{
		CounterMetrics: counterMetrics,
		GaugeMetrics:   gaugeMetrics,
	}
	tmpl, err := template.New("metrics").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Metric List</title>
		</head>
		<body>
			<h1>Metric List</h1>
			<h2>Counter Metrics</h2>
			<ul>
				{{range $name, $value := .CounterMetrics}}
					<li>{{$name}}: {{$value}}</li>
				{{end}}
			</ul>
			<h2>Gauge Metrics</h2>
			<ul>
				{{range $name, $value := .GaugeMetrics}}
					<li>{{$name}}: {{$value}}</li>
				{{end}}
			</ul>
		</body>
		</html>
	`)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/html")
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}
func HandlerReadMetric (c *gin.Context) {
	name := c.Param("name")
	metricType := c.Param("type")
	var value interface{}
	if name == "" {c.AbortWithStatus(http.StatusNotFound); return}
	var ok bool
	switch metricType{
	case "gauge": 
		_, gaugeMetrics := db.GetMetrics()
		value, ok = gaugeMetrics[name]
	case "counter": 
		counterMetrics, _ := db.GetMetrics()
		value, ok = counterMetrics[name]
	default: ok = false
	}
	if !ok {c.AbortWithStatus(http.StatusNotFound); return}
	c.String(http.StatusOK, "%v", value)
}
func HandlerWriteMetric(c *gin.Context) {
	name := c.Param("name")
	value := c.Param("value")
	metricType := c.Param("type")
	badReq := 0

	if name == "" {c.AbortWithStatus(http.StatusNotFound); return}
	switch metricType {
	case "counter":
		if val, err := strconv.ParseInt(value, 10, 64); err == nil {
			db.AddCounterMetric(name, val)
		} else {badReq = 1}
	case "gauge":
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			db.AddGaugeMetric(name, val)
		} else {badReq = 1}
	default: badReq = 1
	}
	if badReq == 1 {c.AbortWithStatus(http.StatusBadRequest); return}

	c.Status(http.StatusOK)
}

func main() {
	router := gin.Default()

	router.POST("/update/:type/:name/:value", HandlerWriteMetric)
	router.GET("/value/:type/:name", HandlerReadMetric)
	router.GET("/", HandlerListMetrics)

	parseFlags()
	if err := router.Run(flagRunAddr); err != nil {panic(err)}
}

