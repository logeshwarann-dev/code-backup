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

	router.POST("/api/v1/changeOPS", HandleOPSRate)
	router.PUT("/api/v1/updateRecords", UpdateRecords)
	router.DELETE("/api/v1/deleteOrders", DeleteOrders)
	router.DELETE("/api/v1/deleteRecords", DeleteRecords)
	router.POST("/api/v1/generatePattern", GeneratePattern)
	router.POST("/api/v1/addSlidingPriceRange", AddSlidingPrice)
	// ==========Simulator App functions=========
	/*
		router.POST("api/v1/setConfig", SetConfig)
		router.POST("api/v1/setRecords", SetRecords)
		router.POST("/api/v1/health-check", HealthCheck)
		router.POST("/api/v1/orderPumpingStatus", CheckOrderPumpingStatus)
	*/

	return router
}
