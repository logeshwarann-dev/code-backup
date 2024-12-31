package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type RequestBody struct {
	Message string `json:"message"`
}

func main() {
	router := gin.Default()
	router.POST("/send", func(c *gin.Context) {
		var reqBody RequestBody
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		requestBody, err := json.Marshal(reqBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp, err := http.Post("http://service-b:8082/receive", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		fmt.Printf("Response from Service B: %d\n", resp.StatusCode)
		c.JSON(http.StatusOK, gin.H{"message": "Request sent to Service B"})
	})

	// Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	router.POST("/scale-service-b", func(c *gin.Context) {
		var request struct {
			ReplicaCount int `json:"replica_count"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Scale Service B
		namespace := "default" // Update as needed
		deploymentName := "service-b"

		scaleErr := scaleDeployment(clientset, namespace, deploymentName, request.ReplicaCount)
		if scaleErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": scaleErr.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Service B scaled successfully"})
	})

	router.Run(":8081")
}

func scaleDeployment(clientset *kubernetes.Clientset, namespace, deploymentName string, replicas int) error {
	scale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = int32(replicas)

	_, err = clientset.AppsV1().Deployments(namespace).UpdateScale(context.TODO(), deploymentName, scale, metav1.UpdateOptions{})
	return err
}
