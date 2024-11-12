package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)

	responseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_status_total",
			Help: "Total number of HTTP responses by status code",
		},
		[]string{"status"},
	)

	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path"},
	)
)

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Invoke the next handler in the middleware chain
		err := next(c)

		// Check for any errors
		if err != nil {
			log.Printf("Request failed: %s %s returned error: %s",
				c.Request().Method, c.Request().URL.Path, err.Error())
			return err
		}

		// Check the status of the response
		if c.Response().Status >= http.StatusBadRequest {
			log.Printf("Request failed: %s %s returned status %d",
				c.Request().Method, c.Request().URL.Path, c.Response().Status)
		}

		return nil
	}
}

// Middleware for custom Prometheus metrics
func prometheusMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Path()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path)) // Start the timer

		// Call the next handler
		err := next(c)

		// Capture the response status code
		statusCode := c.Response().Status

		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc() // Increment status code counter
		totalRequests.WithLabelValues(path).Inc()                      // Increment total requests counter

		timer.ObserveDuration() // Stop the timer and observe the duration

		return err
	}
}

func main() {
	prometheus.MustRegister(httpDuration, responseStatus, totalRequests)

	e := echo.New()

	//e.Use(LoggerMiddleware) // Custom logging middleware
	e.Use(prometheusMiddleware)

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := e.Start(":8080"); err != nil {
			fmt.Println("Error starting server:", err)
		}
	}()

	<-done

	fmt.Println("Server shutting down gracefully...")
}
