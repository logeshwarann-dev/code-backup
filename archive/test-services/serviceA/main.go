package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RequestBody struct {
	Message string `json:"message"`
}

func main() {
	router := gin.Default()
	router.POST("/send", func(c *gin.Context) {
		var reqBody RequestBody
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		requestBody, err := json.Marshal(reqBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp, err := http.Post("http://service-b:8082/receive", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Response from Service B: %d\n", resp.StatusCode)
		c.JSON(http.StatusOK, gin.H{"message": "Request sent to Service B"})
	})
	router.Run(":8081")
}
