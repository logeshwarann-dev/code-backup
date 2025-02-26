package api

import (
	"fmt"
	"market_replay/src/core"
	"market_replay/src/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// func ConnectToRedisAndLoadData(context *gin.Context) {
// 	if err := utils.ConnectToSecondaryRedisPod(); err != nil {
// 		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
// 		return
// 	}
// 	context.JSON(http.StatusOK, gin.H{"message": "Redis pod Connection is successful and data has been loaded."})

// }

func StartPumping(context *gin.Context) {

	go utils.ConnectToSecondaryRedisPod()
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
