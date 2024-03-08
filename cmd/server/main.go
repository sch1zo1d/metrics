package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"

	. "github.com/sch1zo1d/metrics/internal/constant"
	"github.com/sch1zo1d/metrics/internal/logger"
)

type MemStorage struct {
	Data struct {
		gauge   GaugeMetric
		counter CounterMetric
	}
	mu sync.Mutex
}

type Storage interface {
	AddCounterMetric(name string, value int64) Counter
	AddGaugeMetric(name string, value float64) Gauge
	GetMetrics() (CounterMetric, GaugeMetric)
}

const (
	flagLogLevel = "info"
)

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

func (db *MemStorage) AddCounterMetric(name string, value int64) Counter {
	db.mu.Lock()
	db.Data.counter[name] += Counter(value)
	db.mu.Unlock()
	return db.Data.counter[name]
}

func (db *MemStorage) AddGaugeMetric(name string, value float64) Gauge {
	db.mu.Lock()
	db.Data.gauge[name] = Gauge(value)
	db.mu.Unlock()
	return db.Data.gauge[name]
}

func (db *MemStorage) GetMetrics() (CounterMetric, GaugeMetric) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.Data.counter, db.Data.gauge
}

func SerealizeJSON(c *gin.Context, metric *Metrics) {
	var buf bytes.Buffer
	if c.ContentType() != "application/json" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

func HandlerListMetrics(c *gin.Context) {
	counterMetrics, gaugeMetrics := db.GetMetrics()

	data := struct {
		CounterMetrics map[string]Counter
		GaugeMetrics   map[string]Gauge
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

// value/
func HandlerReadMetric(c *gin.Context) {
	var metric Metrics
	SerealizeJSON(c, &metric)

	// что будет, если человек отправит метрику с двумя значениями?
	// а если с заполненным одним, но другим типом?
	// выведет то, что он отправил

	name := metric.ID
	metricType := metric.MType
	var value interface{}
	if name == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var ok bool
	switch metricType {
	case CounterS:
		counterMetrics, _ := db.GetMetrics()
		value, ok = counterMetrics[name]
		*metric.Delta = value.(int64)
	case GaugeS:
		_, gaugeMetrics := db.GetMetrics()
		value, ok = gaugeMetrics[name]
		*metric.Value = value.(float64)
	default:
		ok = false
	}
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, metric)
}

// update/
func HandlerWriteMetric(c *gin.Context) {
	var metric Metrics

	SerealizeJSON(c, &metric)

	if metric.ID == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	log.Println(metric)
	// что будет, если человек отправит метрику с пустым значением?
	switch metric.MType {
	case CounterS:
		*metric.Delta = int64(db.AddCounterMetric(metric.ID, *metric.Delta))
	case GaugeS:
		*metric.Value = float64(db.AddGaugeMetric(metric.ID, *metric.Value))
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, metric)
}

func initRouter() (router *gin.Engine) {
	if err := logger.Initialize(flagLogLevel); err != nil {
		log.Panic("Can't init router")
	}
	router = gin.New()
	router.Use(logger.Logger(logger.Log))

	router.POST("/update/", HandlerWriteMetric)
	router.GET("/value/", HandlerReadMetric)
	router.GET("/", HandlerListMetrics)
	return router
}

func main() {
	router := initRouter()

	parseFlags()
	if err := router.Run(flagRunAddr); err != nil {
		panic(err)
	}
}
