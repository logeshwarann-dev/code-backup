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
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	server := gin.Default()

	server.POST("/multiredis", connectToRedis)

	server.Run(":8080")

}

func connectToRedis(context *gin.Context) {
	if err := RedisConnection(); err != nil {
		context.JSON(http.StatusOK, gin.H{"message": "Error in Redis operation."})

	}
	context.JSON(http.StatusOK, gin.H{"message": "Redis Connection is successful!"})
	return
}
