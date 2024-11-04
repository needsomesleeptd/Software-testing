package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func main() {
	r := gin.Default()

	// Initialize Prometheus middleware
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)
	// Metrics endpoint
	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	// Setup signal handling for graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start the server on port 8080 in a goroutine
	go func() {
		if err := r.Run("localhost:8080"); err != nil {
			fmt.Println("Error starting server:", err)
		}
	}()

	// Wait for termination signal
	<-done

	fmt.Println("Server shutting down gracefully...")
}
