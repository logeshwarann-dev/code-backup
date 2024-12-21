// package main

// import (
// 	"fmt"
// )

// func main() {
// 	fmt.Println("STARTED!!!!")
// 	RedisConnection()

// 	fmt.Println("Press Enter to exit...")
// 	fmt.Scanln()

// }

package main

import (
	"multi-redis/src/api"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	router.POST("/api/v1/data-processor/redis/store-data-in-secondary", api.ConnectToRedis)

	router.POST("/api/v1/data-processor/trigger-sender", api.TriggerSender)

	router.POST("/api/v1/data-processor/get-pod-health", api.GetHealthStatus)

	router.Run(":8080")

}
