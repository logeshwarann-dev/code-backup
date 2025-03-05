package main

import (
	"fmt"
)

func main() {
	// if err := logger.Init(); err != nil {
	// 	log.Fatalf("Failed to initialize logger: %v", err)
	// }
	fmt.Println("[START] GO RUNN!!!!!!!!!!!!")

	SessionIdSets = make(map[string][]Member)

	// go Start()
	router := SetupRouter()
	router.Run(":8080")

}
