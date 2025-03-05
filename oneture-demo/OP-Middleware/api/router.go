package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Adjust the allowed origins as needed
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.POST("/opMiddleware/v1/changeOPS", HandleOPSRate)

	router.PUT("/opMiddleware/v1/updateRecords", UpdateRecords)

	router.DELETE("/opMiddleware/v1/deleteOrders", DeleteOrders)

	router.DELETE("/opMiddleware/v1/deleteRecords", DeleteRecords)

	router.GET("/opMiddleware/v1/getRecords", GetRecords)

	router.POST("/opMiddleware/v1/setConfig", SetConfig)

	router.GET("/opMiddleware/v1/getConfig", GetConfig)

	router.GET("/events", SseHandler)

	// router.GET("/opMiddleware/v1/stopOrderPumping",StopOrderPumping)

	router.POST("/opMiddleware/v1/generatePattern", GeneratePattern)

	router.POST("/opMiddleware/v1/addSlidingPriceRange", AddSlidingPrice)

	router.POST("/opMiddleware/v1/addPods", AddPods)

	router.DELETE("/opMiddleware/v1/deletePods", DeletePods)

	return router
}
