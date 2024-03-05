package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"text/template"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)
func TestHandlerListMetrics(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.GET("/", HandlerListMetrics)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerReadMetric(t *testing.T) {
	req, _ := http.NewRequest("GET", "/update/gauge/metric_name", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.GET("/update/:type/:name", HandlerReadMetric)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "0", w.Body.String())
}

func TestHandlerWriteMetric(t *testing.T) {
	req, _ := http.NewRequest("POST", "/update/counter/metric_name/10", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.POST("/update/:type/:name/:value", HandlerWriteMetric)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
type MemStorage struct {
	Data struct {
		gauge   GaugeMetric
		counter CounterMetric
	}
	mu sync.Mutex
}

// type gauge float64
// type counter int64

type GaugeMetric map[string]gauge
type CounterMetric map[string]counter

type Storage interface {
	AddCounterMetric(name string, value int64)
	AddGaugeMetric(name string, value float64)
	GetMetrics() (CounterMetric, GaugeMetric)
}

var (

	database Storage = &MemStorage{
		Data: struct {
			gauge   GaugeMetric
			counter CounterMetric
		}{
			gauge:   make(GaugeMetric),
			counter: make(CounterMetric),
		},
	}
)

func (database *MemStorage) AddCounterMetric(name string, value int64) {
	database.mu.Lock()
	database.Data.counter[name] += counter(value)
	database.mu.Unlock()
}

func (database *MemStorage) AddGaugeMetric(name string, value float64) {
	database.mu.Lock()
	database.Data.gauge[name] = gauge(value)
	database.mu.Unlock()
}

func (database *MemStorage) GetMetrics() (CounterMetric, GaugeMetric) {
	database.mu.Lock()
	defer database.mu.Unlock()
	return database.Data.counter, database.Data.gauge
}


func HandlerListMetrics(c *gin.Context) {
	counterMetrics, gaugeMetrics := database.GetMetrics()

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
	switch metricType{
	case "gauge": 
		_, gaugeMetrics := database.GetMetrics()
		value = gaugeMetrics[name]
	case "counter": 
		counterMetrics, _ := database.GetMetrics()
		value = counterMetrics[name]
	default: c.AbortWithStatus(http.StatusNotFound); return
	}
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
			database.AddCounterMetric(name, val)
		} else {badReq = 1}
	case "gauge":
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			database.AddGaugeMetric(name, val)
		} else {badReq = 1}
	default: badReq = 1
	}
	if badReq == 1 {c.AbortWithStatus(http.StatusBadRequest); return}

	c.Status(http.StatusOK)
}
