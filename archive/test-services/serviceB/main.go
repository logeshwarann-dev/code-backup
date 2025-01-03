package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.POST("/receive", func(c *gin.Context) {
		var reqBody map[string]interface{}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Printf("Received message from Service A: %+v\n", reqBody)
		c.JSON(http.StatusOK, gin.H{"message": "Message received"})
	})
	router.Run(":8082")
}
