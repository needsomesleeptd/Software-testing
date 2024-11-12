package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

// Middleware for custom Prometheus metrics
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next() // Call the next handler

		// Measure duration
		duration := time.Since(start).Seconds()
		path := c.Request.URL.Path

		// Capture metrics
		httpDuration.WithLabelValues(path).Observe(duration)
		statusCode := strconv.Itoa(c.Writer.Status())
		responseStatus.WithLabelValues(statusCode).Inc()
		totalRequests.WithLabelValues(path).Inc()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Log the status of the request

		log.Printf("Request failed: %s %s returned status %d",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status())

	}
}

func main() {
	// Register Prometheus metrics
	prometheus.MustRegister(httpDuration, responseStatus, totalRequests)

	// Create a new Gin router
	r := gin.Default()

	// Use the custo	m middleware for metrics
	r.Use(prometheusMiddleware())
	//r.Use(Logger())

	// Define the /metrics endpoint for Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Define a root endpoint
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		if err := r.Run(":8080"); err != nil { // listens on port 8080
			fmt.Println("Error starting server:", err)
		}
	}()

	// Wait for termination signal
	<-done
	fmt.Println("Server shutting down gracefully...")
}
