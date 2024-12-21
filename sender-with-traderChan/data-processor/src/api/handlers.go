package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	utils "multi-redis/src/utils"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func ConnectToRedis(context *gin.Context) {
	if err := utils.RedisConnection(); err != nil {
		context.JSON(http.StatusOK, gin.H{"message": "Error in Redis operation."})

	}
	context.JSON(http.StatusOK, gin.H{"message": "Redis Connection is successful!"})
	// return
}

func TriggerSender(context *gin.Context) {
	responses := fanOutRequestsToSenders("/api/v1/sender/start-pumping", "Start Pumping!!")
	context.JSON(http.StatusOK, gin.H{"responses": responses})
}

func GetHealthStatus(context *gin.Context) {
	responses := fanOutRequestsToSenders("/api/v1/sender/health-check", "Health Check")
	context.JSON(http.StatusOK, gin.H{"responses": responses})
}

func fanOutRequestsToSenders(endpoint string, reqMsg string) []map[string]interface{} {
	podsList := getSecondaryRedisPodAddr("client")
	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	// Iterate over pods
	for _, pod := range podsList {
		senderPodIP := pod.Status.PodIP
		senderAddr := fmt.Sprintf("http://%s:8090%s", senderPodIP, endpoint)

		wg.Add(1) // Increment the wait group counter
		go func(addr string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			resp, err := http.Post(addr, "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"message": %s}`, reqMsg))))
			if err != nil {
				mu.Lock()
				responses = append(responses, map[string]interface{}{
					"address": addr,
					"error":   err.Error(),
				})
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			mu.Lock()
			responses = append(responses, map[string]interface{}{
				"address":  addr,
				"response": string(body),
			})
			mu.Unlock()
		}(senderAddr)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Return the consolidated responses
	return responses
}

func getSecondaryRedisPodAddr(appLabel string) []v1.Pod {

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating Kubernetes config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	// Get list of secondary Redis pods
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", appLabel),
	})
	if err != nil {
		log.Fatalf("Error getting client pods: %v", err)
	}

	return pods.Items

}
