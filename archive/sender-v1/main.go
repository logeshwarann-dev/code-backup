// package main

// import (
// 	"fmt"
// )

// var mainChannel = make(chan []string, MAIN_CHANNEL_LENGTH)

// func main() {
// 	fmt.Println("GO RUNN!!!!!!!!!!!!")

//

// 	fmt.Println("Press Enter to exit...")
// 	fmt.Scanln()
// }

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var mainChannel = make(chan []string, MAIN_CHANNEL_LENGTH)

func main() {

	server := gin.Default()

	server.POST("/go-sender", startPumping)

	server.Run(":8090")

}

func startPumping(context *gin.Context) {

	go Start()

	context.JSON(http.StatusOK, gin.H{"message": "Order Pumping has started!"})
	return

}
