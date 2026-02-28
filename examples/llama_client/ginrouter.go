package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatRoomMessage matches the payload sent by the Java Spring app.
// Spring's toString() format: ChatRoomMessage[sender=You, content=hello, type=CHAT]
// When sent as JSON from Spring it looks like:
//
//	{"sender":"You","content":"hello","type":"CHAT"}
type ChatRoomMessage struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

// ChatReply is the JSON response returned to the Spring client.
type ChatReply struct {
	Reply string `json:"reply"`
}

// StartGinRouter registers routes and starts the Gin HTTP server.
// It blocks until the server stops.
func StartGinRouter(s *LlamaServer, addr string) {
	r := gin.Default()

	// POST /chat  — accepts ChatRoomMessage, returns ChatReply
	r.POST("/chat", func(c *gin.Context) {
		var msg ChatRoomMessage
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
			return
		}

		if msg.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content must not be empty"})
			return
		}

		reply, err := s.Ask(c.Request.Context(), msg.Content)
		if err != nil {
			log.Printf("[Chat] error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ChatReply{Reply: reply})
	})

	log.Printf("Gin HTTP server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Gin server failed: %v", err)
	}
}
