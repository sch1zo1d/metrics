package logger

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Log будет доступен всему коду как синглтон.
// Никакой код навыка, кроме функции InitLogger, не должен модифицировать эту переменную.
// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
    // преобразуем текстовый уровень логирования в zap.AtomicLevel
    lvl, err := zap.ParseAtomicLevel(level)
    if err != nil {
        return err
    }
    // создаём новую конфигурацию логера
    cfg := zap.NewProductionConfig()
    // устанавливаем уровень
    cfg.Level = lvl
    // создаём логер на основе конфигурации
    zl, err := cfg.Build()
    if err != nil {
        return err
    }
    // устанавливаем синглтон
    Log = zl
    return nil
}


type (
    // берём структуру для хранения сведений об ответе
    responseData struct {
        status int
        size int
    }

    // добавляем реализацию http.ResponseWriter
    loggingResponseWriter struct {
        http.ResponseWriter
        responseData *responseData
    }
)
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
    // записываем ответ, используя оригинальный http.ResponseWriter
    size, err := r.ResponseWriter.Write(b) 
    r.responseData.size += size // захватываем размер
    return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
    // записываем код статуса, используя оригинальный http.ResponseWriter
    r.ResponseWriter.WriteHeader(statusCode) 
    r.responseData.status = statusCode // захватываем код статуса
} 

// middleware-логер для HTTP-запросов.
func WithLogging(h http.HandlerFunc) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData {
            status: 0,
            size: 0,
        }
        lw := loggingResponseWriter {
            ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
            responseData: responseData,
        }
		start := time.Now()

        h(&lw, r)

		duration := time.Since(start)

		Log.Info("HTTP request",
			zap.String("URI", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("duration", duration.String()),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)

    })
}
func Logger(*zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// responseData := &responseData {
		//     status: 0,
		//     size: 0,
		// }
		// lw := loggingResponseWriter {
		//     ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		//     responseData: responseData,
		// }

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		Log.Info("HTTP request",
			zap.String("URI", c.Request.RequestURI),
			zap.String("method", c.Request.Method),
			zap.String("duration", duration.String()),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)

	}
}