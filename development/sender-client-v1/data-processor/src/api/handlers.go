package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	utils "multi-redis/src/utils"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type HistoryFilePath struct {
	FilePath []string `json:"historyfile_path"`
}

// ===========HANDLER FUNCTIONS==============

func TransferOrderDataToRedisPods(context *gin.Context) {

	if err := utils.RedisConnection(); err != nil {
		context.JSON(http.StatusOK, gin.H{"message": "Error in Redis operation."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Redis Connection is successful!"})
	// return
}

func InitiateRedisPodConnWithSender(context *gin.Context) {
	reqBody := `{"message": "Connect To Redis Pod!"}`
	responses := fanOutRequestsToPods("/api/v2/sender/connectToRedis", reqBody, "client", "8090")
	context.JSON(http.StatusOK, gin.H{"responses": responses})
}

func TriggerSender(context *gin.Context) {
	reqBody := `{"message": "Start Pumping"}`
	responses := fanOutRequestsToPods("/api/v1/sender/start-pumping", reqBody, "client", "8090")
	context.JSON(http.StatusOK, gin.H{"responses": responses})
}

func GetSenderHealthStatus(context *gin.Context) {
	reqBody := `{"message": "Health Check"}`
	responses := fanOutRequestsToPods("/api/v1/sender/health-check", reqBody, "client", "8090")
	context.JSON(http.StatusOK, gin.H{"responses": responses})
}

func StartDataProcessing(context *gin.Context) {

	historyFilesPath := HistoryFilePath{}

	if err := context.ShouldBindJSON(&historyFilesPath); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	metricRes, err := TriggerMetricGenerator(historyFilesPath)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": metricRes})
		return
	}
	fileParserRes, err := ScaleFileParserPods(historyFilesPath)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": fileParserRes})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Data Processing has been started!"})

}

func TriggerMetricGenerator(historyFilePath HistoryFilePath) (string, error) {

	requestBody, err := json.Marshal(historyFilePath)
	if err != nil {
		return "Marshalling error", err
	}

	resp, err := http.Post("http://metrics-generator-service:5001/api/v1/metric-gen/store-metrics-in-redis", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "Error in sending request to Metric Generator", err
	}
	defer resp.Body.Close()

	metricGenResponse := fmt.Sprintf("Response from Metric Generator: %d\n", resp.StatusCode)
	return metricGenResponse, nil

}

func ScaleFileParserPods(historyFilePath HistoryFilePath) (string, error) {

	totalFiles := len(historyFilePath.FilePath)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Scale file parser pods
	namespace := "default"
	statefulSet := "fileparser"

	scaleErr := utils.ScaleStatefulSet(clientset, namespace, statefulSet, totalFiles)
	if scaleErr != nil {
		return scaleErr.Error(), scaleErr
	}

	for try := range 5 {
		time.Sleep(5 * time.Second)
		reqBody := `{"message": "Health Check"}`
		responses := fanOutRequestsToPods("/api/v1/fileparser/health-check", reqBody, "fileparser", "5000")
		if len(responses) == totalFiles {
			fmt.Printf("%v fileparser pods have been created and active.\n", totalFiles)
			break
		}
		fmt.Printf("[%d/5] Retrying Health Check for fileparser pods.\n", try)
	}

	for _, filePathValue := range historyFilePath.FilePath {
		reqBody := fmt.Sprintf(`{"file_path": %s}`, filePathValue)
		fanOutRequestsToPods("/api/v1/fileparser/process-file", reqBody, "fileparser", "5000")
	}

	return "File Parser pods are scaled successfully", nil

}

// ============HELPER FUNCTIONS=================

func fanOutRequestsToPods(endpoint string, reqMsg string, appLabel string, servicePort string) []map[string]interface{} {
	podsList := getPodAddr(appLabel)
	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	// Iterate over pods
	for _, pod := range podsList {
		podIP := pod.Status.PodIP
		podAddr := fmt.Sprintf("http://%s:%s%s", podIP, servicePort, endpoint)

		wg.Add(1) // Increment the wait group counter
		go func(addr string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			resp, err := http.Post(addr, "application/json", bytes.NewBuffer([]byte(reqMsg)))
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
		}(podAddr)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Return the consolidated responses
	return responses
}

func getPodAddr(appLabel string) []v1.Pod {

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

// =========ARCHIVE===========
// func ScaleRedisAndSenderPods(targetCount int /*context *gin.Context*/) error {
// 	// var request struct {
// 	// 	ReplicaCount int `json:"replica_count"`
// 	// }
// 	// if err := context.ShouldBindJSON(&request); err != nil {
// 	// 	context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 	// 	return
// 	// }
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		log.Fatalf("Error creating in-cluster config: %v", err)
// 	}
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		log.Fatalf("Error creating Kubernetes client: %v", err)
// 	}

// 	// Scale secondary redis and sender clients
// 	namespace := "default"
// 	statefulSets := []string{"secondary-redis", "client"}

// 	for _, statefulSet := range statefulSets {
// 		scaleErr := scaleStatefulSet(clientset, namespace, statefulSet, targetCount)
// 		if scaleErr != nil {
// 			return fmt.Errorf("pod Scaling Error: %v", scaleErr.Error())
// 		}
// 	}

// 	return nil
// 	// context.JSON(http.StatusOK, gin.H{"message": "Redis and Trader Client pods are scaled successfully"})

// }
