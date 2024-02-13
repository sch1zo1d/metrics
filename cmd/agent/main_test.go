package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "github.com/sch1zo1d/metrics/cmd/server"
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

// func TestGatherMetrics(t *testing.T) {
// 	stopCh := make(chan struct{})
// 	defer close(stopCh)

// 	go gatherMetrics()

// 	time.Sleep(5 * time.Second)

// 	dbMux.RLock()
// 	defer dbMux.RUnlock()

// 	assert.NotEmpty(t, db.gauge["Alloc"])
// 	assert.NotEmpty(t, db.gauge["BuckHashSys"])

// 	stopCh <- struct{}{}
// }

// func TestSendMetrics(t *testing.T) {
// 	stopCh := make(chan struct{})
// 	defer close(stopCh)

// 	go sendMetrics()

// 	time.Sleep(5 * time.Second)


// 	stopCh <- struct{}{}
// }
