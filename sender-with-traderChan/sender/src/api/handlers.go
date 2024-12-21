package api

import (
	"fmt"
	"market_replay/src/core"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func StartPumping(context *gin.Context) {

	go core.Start()
	context.JSON(http.StatusOK, gin.H{"message": "Order Pumping has started!"})
	return

}

func GetHealthStatus(context *gin.Context) {
	podName := os.Getenv("POD_NAME")
	statusMsg := fmt.Sprintf("%s is active", podName)
	context.JSON(http.StatusOK, gin.H{"pod_status": statusMsg})
	return
}
