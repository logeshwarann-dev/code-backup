package main

import (
	"fmt"

	"op-middleware/api"
	static "op-middleware/static"
)

// logger "op-middleware/logging"

func main() {
	// if err := logger.Init(); err != nil {
	// 	log.Fatalf("Failed to initialize logger: %v", err)
	// }
	fmt.Println("[START] GO RUNN!!!!!!!!!!!!")

	static.SenderPodIPMap = make(map[string]string)

	static.EventChannel = make(chan static.SseData, 1000)

	static.TraderPodIPMap = make(map[string]string)

	static.LabSessionIdSets = make(map[string][]static.Trader)
	static.ProdSessionIdSets = make(map[string][]static.Trader)
	static.SimSessionIdSets = make(map[string][]static.Trader)
	static.MemberPodMap = make(map[string]string)
	static.PeakPodIPMap = make(map[string]string)

	static.ConfigDetails.ThrottleValue = make(map[string]int)

	static.CreateSessionIds()

	router := api.SetupRouter()
	router.Run(":8090")
}
