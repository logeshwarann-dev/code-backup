package main

import (
	"market_replay/src/api"
	"market_replay/src/static"
)

func main() {
	// if err := logger.Init(); err != nil {
	// 	log.Fatalf("Failed to initialize logger: %v\n", err)
	// }
	// utils.Printf(static.LOG_FLAG, "[START] GO RUNN!!!!!!!!!!!!")

	static.SessionIdSets = make(map[string][]static.Member)
	static.LOG_FLAG = true
	// go Start()
	router := api.SetupRouter()
	router.Run(":8080")

}
