// ==========================
// cmd/server/main.go (Updated)
// ==========================
package main

import (
	"awesomeProject/internal/config"
	"awesomeProject/internal/handlers"
	"awesomeProject/internal/services"
	ws "awesomeProject/internal/websocket"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate origin properly
	},
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize services
	authService := services.NewAuthService(cfg)
	b2cService := services.NewB2CService(cfg, authService)
	hub := ws.NewHub()
	go hub.Run()

	// Initialize handlers
	stkHandler := handlers.NewSTKHandler(cfg, authService, hub)
	b2cHandler := handlers.NewB2CHandler(cfg, b2cService, hub)

	// Setup Gin router
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// WebSocket endpoint
	r.GET("/ws/payments", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}

		client := ws.NewClient(hub, conn)
		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	})

	// STK Push routes
	stk := r.Group("/api/v1/stk")
	{
		stk.POST("/initiate", stkHandler.InitiateSTKPush)
		stk.POST("/callback", stkHandler.STKPushCallback)
	}

	// B2C routes
	b2c := r.Group("/api/v1/b2c")
	{
		b2c.POST("/payment", b2cHandler.InitiatePayment)
		b2c.POST("/result", b2cHandler.HandleCallback)
		b2c.POST("/timeout", b2cHandler.HandleTimeout)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"clients": len(hub.GetClients()),
		})
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("Server is starting on %s", addr)
	log.Printf("WebSocket endpoint: ws://%s/ws/payments", addr)
	log.Printf("B2C Payment endpoint: POST http://%s/b2c/payment", addr)
	log.Printf("B2C Result callback: POST http://%s/b2c/result", addr)
	log.Printf("B2C Timeout callback: POST http://%s/b2c/timeout", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
