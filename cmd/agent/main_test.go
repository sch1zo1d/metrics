package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/sch1zo1d/metrics/cmd/server"

)
func TestHandlerListMetrics(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.GET("/", server.HandlerListMetrics)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerReadMetric(t *testing.T) {
	req, _ := http.NewRequest("GET", "/update/gauge/metric_name", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.GET("/update/:type/:name", server.HandlerReadMetric)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "0", w.Body.String())
}

func TestHandlerWriteMetric(t *testing.T) {
	req, _ := http.NewRequest("POST", "/update/counter/metric_name/10", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.POST("/update/:type/:name/:value", server.HandlerWriteMetric)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
