package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
	reqP, _ := http.NewRequest("POST", "/update/counter/metric_name/10", nil)
	reqG, _ := http.NewRequest("GET", "/value/gauge/metric_name", nil)
	router := gin.Default()
	w := httptest.NewRecorder()
	router.POST("/update/:type/:name/:value", HandlerWriteMetric)
	router.ServeHTTP(w, reqP)
	router.GET("/value/:type/:name", HandlerReadMetric)
	router.ServeHTTP(w, reqG)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Body.String())
}

func TestHandlerWriteMetric(t *testing.T) {
	req, _ := http.NewRequest("POST", "/update/counter/metric_name/10", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.POST("/update/:type/:name/:value", HandlerWriteMetric)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMainFunction(t *testing.T) {
	go func() {
		main()
	}()
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router := gin.Default()
	router.GET("/", HandlerListMetrics)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// assert.Nil()
}
