package main

import (
	"fmt"
	"market_replay/src/api"
	"market_replay/src/static"
)

func main() {
	// if err := logger.Init(); err != nil {
	// 	log.Fatalf("Failed to initialize logger: %v", err)
	// }
	fmt.Println("[START] GO RUNN!!!!!!!!!!!!")

	static.SessionIdSets = make(map[string][]static.Member)

	// go Start()
	router := api.SetupRouter()
	router.Run(":8080")

}
