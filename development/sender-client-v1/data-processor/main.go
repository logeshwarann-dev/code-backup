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

	router.POST("/api/v1/data-processor/scale-fileparser", api.StartDataProcessing) // Body: { "historyfile_path": ["", ""]}

	// router.POST("/api/v1/data-processor/scale-redis-sender-pods", api.ScaleRedisAndSenderPods) // Body: {"replica_count": int}

	router.POST("/api/v1/data-processor/sender/connectToRedisPod", api.InitiateRedisPodConnWithSender) // Body : {}

	router.POST("/api/v1/data-processor/redis/store-data-in-secondary", api.TransferOrderDataToRedisPods) // Body: {}

	router.POST("/api/v1/data-processor/sender/trigger", api.TriggerSender) // Body: {}

	router.POST("/api/v1/data-processor/get-sender-pod-health", api.GetSenderHealthStatus) // Body: {}

	router.POST("/api/test", api.MapPods)
	//Body: { "pod-id": 1, "throttle": 1000}

	router.Run(":8080")

}
