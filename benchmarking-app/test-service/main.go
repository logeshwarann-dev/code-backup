package main

import "github.com/gin-gonic/gin"

var NodeAndPodMap map[string][]string

// var POD_SET = 1

func main() {
	NodeAndPodMap = make(map[string][]string)

	router := gin.Default()

	router.POST("/api/distribute-pods", DistributePods)
	router.Run(":8085")

}
