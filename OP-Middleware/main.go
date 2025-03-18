package main

import (
	"op-middleware/api"
	static "op-middleware/static"
	"op-middleware/utils"
)

// logger "op-middleware/logging"

func init() {
	utils.CreateK8sClient()
}

func main() {
	// if err := logger.Init(); err != nil {
	// 	log.Fatalf("Failed to initialize logger: %v", err)
	// }

	static.SenderPodIPMap = make(map[string]string)

	static.EventChannel = make(chan static.SseData, 1000)

	static.TraderPodIPMap = make(map[string]string)

	static.LabSessionIdSets = make(map[string][]static.Trader)
	static.ProdSessionIdSets = make(map[string][]static.Trader)
	static.SimSessionIdSets = make(map[string][]static.Trader)
	static.MemberPodMap = make(map[string]string)
	static.SyncPeakPodIPMap = make(map[string]string)
	static.AsyncIOCPeakPodIPMap = make(map[string]string)
	static.AsyncEMCPeakPodIPMap = make(map[string]string)

	static.ConfigDetails.ThrottleValue = make(map[string]int)

	static.CreateSessionIds()

	router := api.SetupRouter()
	router.Run(":8090")
}
