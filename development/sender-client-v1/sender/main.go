// package main

// import (
// 	"fmt"
// )

// var mainChannel = make(chan []string, MAIN_CHANNEL_LENGTH)

// func main() {
// 	fmt.Println("GO RUNN!!!!!!!!!!!!")

// 	// Start()

// 	fmt.Println("Press Enter to exit...")
// 	fmt.Scanln()
// }

package main

import (
	api "market_replay/src/api"

	"github.com/gin-gonic/gin"
)

func main() {

	server := gin.Default()

	server.POST("/api/v2/sender/connectToRedis", api.ConnectToRedisAndLoadData)

	server.POST("/api/v1/sender/start-pumping", api.StartPumping)

	server.POST("/api/v1/sender/health-check", api.GetHealthStatus)

	server.Run(":8090")

}
