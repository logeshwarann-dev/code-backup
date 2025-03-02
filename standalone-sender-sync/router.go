package main

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
	router.DELETE("/api/v1/DeleteOrders", DeleteOrders)
	router.DELETE("/api/v1/DeleteRecords", DeleteRecords)
	router.POST("/api/v1/generatePattern", GeneratePattern)
	router.POST("/api/v1/addSlidingPriceRange", AddSlidingPrice)

	return router
}
