package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	static "op-middleware/static"
	"op-middleware/utils"
	"strings"

	// logger "op-middleware/logging"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

func HandleOPSRate(context *gin.Context) {
	var opsRateData static.OPSRate
	if err := context.ShouldBindJSON(&opsRateData); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body" + err.Error(),
		})
		return
	}

	if opsRateData.Type != 0 && opsRateData.Type != 1 {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid type value",
		})
		return
	}

	if opsRateData.Throttle < 0 {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Throttle value must be a non-negative integer",
		})
		return
	}

	if opsRateData.Type == 0 {
		for _, pod := range opsRateData.Pods {
			if _, ok := static.SenderPodIPMap[static.SENDER_POD_LABEL+"-"+strconv.Itoa(pod)]; !ok {
				context.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid pod number" + strconv.Itoa(pod),
				})
				return
			}
		}

		if len(opsRateData.Pods) == 0 {
			opsRateData.Pods = utils.GetActivePodIds("mock-sender-")
		}
		opsRateData.Throttle /= (static.SENDER_CONNECTION_COUNT)
	} else {
		for _, pod := range opsRateData.Pods {
			if _, ok := static.TraderPodIPMap[static.TRADER_POD_LABEL+"-"+strconv.Itoa(pod)]; !ok {
				context.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid pod number" + strconv.Itoa(pod),
				})
				return
			}
		}

		if len(opsRateData.Pods) == 0 {
			for i := 0; i < len(static.TraderPodIPMap); i++ {
				opsRateData.Pods = append(opsRateData.Pods, i)
			}
		}

		opsRateData.Throttle /= static.TRADER_CONNECTION_COUNT
	}

	fmt.Println("Received request to change OPS rate")

	context.JSON(http.StatusOK, gin.H{
		"message": "New OPS Received received successfully.",
	})

	if err := utils.SetOPSRateInPods(opsRateData); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error in changing ops rate: %v", err.Error()),
		})
	}

	static.EventChannel <- static.SseData{
		Message: "Change OPS Completed",
		Type:    0,
	}
}

func UpdateRecords(context *gin.Context) {
	var updateRecordsData static.RecordsArray
	if err := context.ShouldBindJSON(&updateRecordsData); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body" + err.Error(),
		})
		return
	}

	for i, record := range updateRecordsData.Records {
		if err := utils.ValidateRecord(record); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation error" + strconv.Itoa(i) + err.Error(),
			})
			return
		}
	}

	fmt.Println("Records received successfully")
	context.JSON(http.StatusOK, gin.H{
		"message": "Records Received successfully",
	})

	for _, record := range updateRecordsData.Records {
		instId := record.InstrumentID
		foundFlag := 0
		for i, r := range static.RECORDS {
			if r.InstrumentID == instId {
				foundFlag = 1
				static.RECORDS[i] = record
				break
			}
		}
		if foundFlag == 0 {
			static.RECORDS = append(static.RECORDS, record)
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	// Iterate over pods
	for podName, podIP := range static.SenderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/updateRecords"

		msg := static.RecordsArray{
			Records: updateRecordsData.Records,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error marshalling request" + err.Error(),
			})
			return
		}

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("PUT", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in updating records for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in updating records for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in updating records for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in updating records for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				mu.Lock()
				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Records updated successfully for " + podName, Type: 0}
				fmt.Println("Records updated successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	for podName, podIP := range static.TraderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/updateRecords"

		msg := static.RecordsArray{
			Records: updateRecordsData.Records,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error marshalling request" + err.Error(),
			})
			return
		}

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("PUT", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in updating records for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in updating records for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in updating records for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in updating records for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				mu.Lock()
				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Records updated successfully for " + podName, Type: 0}
				fmt.Println("Records updated successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	static.EventChannel <- static.SseData{
		Message: "Update Records Completed",
		Type:    0,
	}
}

func DeleteRecords(context *gin.Context) {
	var data static.DeleteRecordStruct
	if err := context.ShouldBindBodyWithJSON(&data); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Request format is not proper"})
		return
	}

	foundFlag := 0

	for i, record := range static.RECORDS {
		if record.InstrumentID == data.InstId {
			foundFlag = 1
			static.RECORDS = append(static.RECORDS[:i], static.RECORDS[i+1:]...)
			break
		}
	}

	if foundFlag == 0 {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "No Instrument Id present as " + strconv.Itoa(data.InstId),
		})
		return
	}

	fmt.Println("Record delete request received successfully")
	context.JSON(http.StatusOK, gin.H{
		"message": "Record will be deleted.",
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	for podName, podIP := range static.SenderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/deleteRecords"

		msg := static.DeleteRecordStruct{
			InstId: data.InstId,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error marshalling request" + err.Error(),
			})
			return
		}

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("DELETE", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting record for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting record for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting record for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting record for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				mu.Lock()
				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Record deleted successfully for " + podName, Type: 0}
				fmt.Println("Record deleted successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	for podName, podIP := range static.TraderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/deleteRecords"

		msg := static.DeleteRecordStruct{
			InstId: data.InstId,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error marshalling request" + err.Error(),
			})
			return
		}

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("DELETE", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting record for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting record for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting record for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting record for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				mu.Lock()

				responseBody := string(body)
				if strings.Contains(responseBody, "error") {
					fmt.Println("[ERR] Error in deleting record in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in deleting record for " + podName, Type: 1}
					continue
				}
				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Record deleted successfully for " + podName, Type: 0}
				fmt.Println("Record deleted successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	static.EventChannel <- static.SseData{
		Message: "Delete Orders Completed",
		Type:    0,
	}
}

func DeleteOrders(context *gin.Context) {
	var deleteOrder static.DeleteOrder

	if err := context.ShouldBindJSON(&deleteOrder); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body" + err.Error(),
		})
		return
	}

	if err := utils.DeleteOrdersValidationCheck(deleteOrder); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation error" + err.Error(),
		})
		return
	}

	fmt.Println("Orders delete request received successfully")
	context.JSON(http.StatusOK, gin.H{
		"message": "Orders delete request received successfully",
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	// Iterate over pods
	for podName, podIP := range static.SenderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/deleteOrders"

		msg := static.DeleteOrder{
			InstrumentID: deleteOrder.InstrumentID,
			ProductID:    deleteOrder.ProductID,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error marshalling request" + err.Error(),
			})
			return
		}

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("DELETE", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting orders for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting orders for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting orders for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting orders for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				mu.Lock()
				responseBody := string(body)
				if strings.Contains(responseBody, "error") {
					fmt.Println("[ERR] Error in deleting orders in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in deleting orders for" + podName, Type: 1}
					continue
				}
				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Orders deleted successfully for " + podName, Type: 0}
				fmt.Println("Orders deleted successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	for podName, podIP := range static.TraderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/deleteOrders"

		msg := static.DeleteOrder{
			InstrumentID: deleteOrder.InstrumentID,
			ProductID:    deleteOrder.ProductID,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error marshalling request" + err.Error(),
			})
			return
		}

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("DELETE", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting orders for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting orders for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in deleting orders for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in deleting orders for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				mu.Lock()

				responseBody := string(body)
				if strings.Contains(responseBody, "error") {
					fmt.Println("[ERR] Error in deleting orders in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in deleting orders for" + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Orders deleted successfully for " + podName, Type: 0}
				fmt.Println("Orders deleted successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	static.EventChannel <- static.SseData{
		Message: "Delete Orders Completed",
		Type:    0,
	}
}

func GetRecords(context *gin.Context) {
	fmt.Println("Records sent successfully")
	context.JSON(http.StatusOK, gin.H{
		"message": "Records Sent successfully",
		"records": static.RECORDS,
	})

}

func SetConfig(context *gin.Context) {

	// static.ConfigData.UniqueClientIdentifier = true

	if err := context.ShouldBindJSON(&static.ConfigData); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "config static.ConfigData not in proper format" + err.Error(),
		})
		return
	}

	if err := utils.ValidateConfig(static.ConfigData); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: err.Error(),
			Type:    1,
		}
		return
	}

	fmt.Println("Config received: ", static.ConfigData)
	static.EventChannel <- static.SseData{
		Message: "Config received. Processing",
		Type:    0,
	}

	static.PROCESS_STATE = 1
	static.TARGETENV = static.ConfigData.TargetEnv

	clientset, err := utils.CreateK8sClient()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create k8s client",
		})
		return
	}

	senderTargetCount := int(math.Ceil(float64(static.ConfigData.TraderCount) / float64(static.SENDER_CONNECTION_COUNT)))

	err = utils.ScaleStatefulSet(clientset, static.K8S_NAMESPACE, static.SENDER_POD_LABEL, senderTargetCount)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while generating sender pods" + err.Error(),
		})
		fmt.Println("Error while generating sender pods" + err.Error())
		return
	}

	err = utils.ScaleStatefulSet(clientset, static.K8S_NAMESPACE, static.TRADER_POD_LABEL, static.TRADER_POD_COUNT)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while generating trader pods" + err.Error(),
		})
		fmt.Println("Error while generating trader pods" + err.Error())
		return
	}

	if err := utils.CreatePodMap(senderTargetCount); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create pod map",
		})
		return
	}

	responses := make([]map[string]interface{}, 0)

	if err := utils.SetConfigInTraderPods(static.ConfigData, static.TraderPodIPMap, responses); err != nil {
		fmt.Printf("Error setting config in sender pods: %v", err.Error())
	}

	static.CalculatedSenderThrottle = static.ConfigData.ThrottleValue / (len(static.SenderPodIPMap) * static.TRADER_CONNECTION_COUNT)

	if err := utils.SetConfigInSenderPods(static.ConfigData, static.SenderPodIPMap, responses); err != nil {
		fmt.Printf("Error setting config in sender pods: %v", err.Error())
	}

	records, err := utils.ReadRecords(static.FilePath)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error reading records" + err.Error(),
		})
		return
	}

	static.RECORDS = append(static.RECORDS, records...)

	if err := utils.SetRecordsInPods(static.SenderPodIPMap, static.RECORDS, responses); err != nil {
		fmt.Printf("Error setting records in sender pods: %v", err.Error())
	}

	if err := utils.SetRecordsInPods(static.TraderPodIPMap, static.RECORDS, responses); err != nil {
		fmt.Printf("Error setting records in trader pods: %v", err.Error())
	}

	static.EventChannel <- static.SseData{
		Message: "Set Config Completed",
		Type:    0,
	}
	context.JSON(http.StatusOK, gin.H{
		"message": "Config applied successfully",
	})

}

func GetConfig(c *gin.Context) {
	fmt.Println("Config fetch request received successfully")
	static.ConfigDetails.ProcessState = static.PROCESS_STATE
	c.JSON(http.StatusOK, gin.H{
		"message": "config Details",
		"config":  static.ConfigDetails,
	})
}

func SseHandler(c *gin.Context) {
	w := c.Writer
	r := c.Request
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case message := <-static.EventChannel:
			fmt.Println("In SSE Handler, ", message)
			Msg, err := json.Marshal(message)
			if err != nil {
				continue
			}

			sseMessage := append([]byte("data: "), Msg...)
			sseMessage = append(sseMessage, '\n', '\n')

			_, _ = w.Write(sseMessage)
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}

func GeneratePattern(c *gin.Context) {
	var generatePatternData static.GraphPattern

	if err := c.ShouldBindJSON(&generatePatternData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "invalid request.",
			Type:    1,
		}
		return
	}

	if err := utils.ValidatePattern(generatePatternData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: fmt.Sprintf("invalid request: %v", err.Error()),
			Type:    1,
		}
		return
	}
	// Peak Generation
	if generatePatternData.Type == 7 {
		var opsRateData static.OPSRate

		// Calculate baseops for sender pods
		baseOps := generatePatternData.Min
		opsRateData.Pods = utils.GetActivePodIds("mock-sender-")
		if len(opsRateData.Pods) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "insufficient sender pods. Please add more pods",
			})
			static.EventChannel <- static.SseData{
				Message: "insufficient sender pods. Please add more pods",
				Type:    1,
			}
			return
		}
		totalSenderPods := len(opsRateData.Pods) - 1 // 1 pod will act as peak generator
		static.CalculatedSenderThrottle = baseOps / (totalSenderPods * static.SENDER_CONNECTION_COUNT)
		opsRateData.Throttle = static.CalculatedSenderThrottle
		opsRateData.Type = 0 //order pumping type

		if len(generatePatternData.Pods) == 0 {
			for peakPodMember, peakPodName := range static.MemberPodMap {
				if strings.Contains(peakPodName, static.TRADER_POD_LABEL) {
					continue
				}
				static.PeakPodIPMap[peakPodName] = static.SenderPodIPMap[peakPodName]
				podId, _ := utils.ExtractPodId(peakPodName)
				generatePatternData.Pods = append(generatePatternData.Pods, podId)
				fmt.Println("In Peak Generation,", "| PeakPodIPMap: ", static.PeakPodIPMap, "| Member: ", peakPodMember, "| Sender OPS RATE: ", opsRateData)
				break
			}
		}

		//starting peak pod
		// var responses []map[string]interface{}
		// if err := utils.SetConfigInPeakPods(static.ConfigData, static.PeakPodIPMap, responses); err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error setting config in peak pods: %v", err.Error())})
		// 	return
		// }

		// if err := utils.SetRecordsInPods(static.PeakPodIPMap, static.RECORDS, responses); err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error setting records in peak pods: %v", err.Error())})
		// 	return
		// }

		// setting base ops in sender pods
		if err := utils.SetOPSRateInPods(opsRateData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("error in changing ops rate: %v", err.Error()),
			})
			return
		}

		// send pattern data to peak pod
		peakOps := generatePatternData.Max - baseOps
		if peakOps > 5000 {
			generatePatternData.Min = 5000
		} else {
			generatePatternData.Min = peakOps / 2
		}
		if generatePatternData.Min < 100 {
			generatePatternData.Min = 100
		}
		// totalPeakPods := len(static.PeakPodIPMap)
		generatePatternData.Max = peakOps /* / (totalPeakPods * static.SENDER_CONNECTION_COUNT) */

		fmt.Println("In Peak Gen, pattern data: ", generatePatternData)

		if err := utils.SendPatternGenerateDataToPods(generatePatternData, static.SENDER_POD_LABEL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("error in sending Pattern to peak pods: %v", err.Error()),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"msg": "Pattern is set in peak pods"})

		static.EventChannel <- static.SseData{
			Message: "Pattern is set in peak pods",
			Type:    0,
		}

		return
	}

	if len(generatePatternData.Pods) == 0 {
		generatePatternData.Pods = utils.GetActivePodIds("mock-sender-")
	}

	fmt.Println("In Pattern gen, sender pod data: ", generatePatternData)

	if err := utils.SendPatternGenerateDataToPods(generatePatternData, static.SENDER_POD_LABEL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error in sending Pattern to sender pods: %v", err.Error()),
		})
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Pattern is set in pods"})

	static.EventChannel <- static.SseData{
		Message: "Pattern is set in pods",
		Type:    0,
	}

}

func AddSlidingPrice(c *gin.Context) {

	var slidingPriceData static.PriceRangeChangeDetails

	if err := c.ShouldBindJSON(&slidingPriceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := utils.ValidatePriceRangeChange(slidingPriceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	if err := utils.ValidatePrices(slidingPriceData.Instruments); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		return
	}

	if err := utils.SetSlidingPriceInPods(slidingPriceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "error in sliding price",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Sliding Price is set in pods"})

	static.EventChannel <- static.SseData{
		Message: "Sliding Price is set in pods",
		Type:    0,
	}

}

func StopOrderPumping(c *gin.Context) {
	fmt.Println("Stop Order Pumping request received successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Stop Order Pumping request received successfully",
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	// Iterate over pods
	for podName, podIP := range static.SenderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/stopOrderPumping"

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			count := static.MaxRetry
			for count > 0 {
				req, err := http.NewRequest("GET", addr, nil)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{
							Message: "Error in stopping order pumping for " + podName + " : " + err.Error(),
							Type:    1,
						}
						fmt.Println("Error in stopping order pumping for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{
							Message: "Error in stopping order pumping for " + podName + " : " + err.Error(),
							Type:    1,
						}
						fmt.Println("Error in stopping order pumping for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				mu.Lock()

				responseBody := string(body)
				if strings.Contains(responseBody, "error") {
					fmt.Println("[ERR] Error in stopping order pumping in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in stopping order pumping for" + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{
					Message: "Order Pumping stopped successfully for " + podName,
					Type:    0,
				}
				fmt.Println("Order Pumping stopped successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	static.PROCESS_STATE = 0
}

func AddPods(c *gin.Context) {
	var addPodPayload static.AddPodCount
	if err := c.ShouldBindJSON(&addPodPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error binding payload" + err.Error()})
		static.EventChannel <- static.SseData{
			Message: "invalid input",
			Type:    1,
		}
		return
	}

	if err := utils.ValidateAddPod(addPodPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: err.Error(),
			Type:    1,
		}
		return
	}

	static.EventChannel <- static.SseData{
		Message: "pod creation in progress",
		Type:    0,
	}

	reqPodCount := addPodPayload.Count
	podsStatusFlag, inactivePodsArr := utils.GetPumpingStatusOfPods()

	availableMemberCount := utils.CheckAvailableMemberIds()
	fmt.Println("Available Member IDs: ", availableMemberCount)
	if availableMemberCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no session ids left. cannot create pods"})
		static.EventChannel <- static.SseData{
			Message: "no session ids left. cannot create pods",
			Type:    1,
		}
		return
	}
	possiblePodCount := availableMemberCount
	if reqPodCount > possiblePodCount {
		c.JSON(http.StatusAccepted, gin.H{"error": fmt.Sprintf("only %d pod(s) will be created", possiblePodCount)})
		static.EventChannel <- static.SseData{
			Message: fmt.Sprintf("only %d pod(s) will be created", possiblePodCount),
			Type:    1,
		}
		reqPodCount = possiblePodCount
	}
	if podsStatusFlag && len(inactivePodsArr) > 0 {
		existingPodStatus, totalPodsStarted := utils.UseExistingPodsInStatefulSet(reqPodCount)

		if existingPodStatus && totalPodsStarted == reqPodCount {
			c.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("%d pod(s) activated", reqPodCount)})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("%d pod(s) created", reqPodCount),
				Type:    0,
			}
			return
		} else if existingPodStatus && totalPodsStarted < reqPodCount {
			reqPodCount -= totalPodsStarted
		}
	}

	if err := utils.UseNewPodsInStatefulSet(reqPodCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Sprintf("error creating new %v pod: %v", reqPodCount, err.Error())})
		static.EventChannel <- static.SseData{
			Message: "error while creating pods",
			Type:    1,
		}
		return
	}

	// fmt.Println("Member Pod Map: ", static.MemberPodMap)
	c.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("%d pods are created", reqPodCount)})

	static.EventChannel <- static.SseData{
		Message: fmt.Sprintf("%d pod(s) created", reqPodCount),
		Type:    0,
	}

}

func DeletePods(c *gin.Context) {

	var deletePodCount static.DeletePodCount
	if err := c.ShouldBindJSON(&deletePodCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error binding payload" + err.Error()})
		static.EventChannel <- static.SseData{
			Message: "error binding payload",
			Type:    1,
		}
		return
	}

	if err := utils.ValidateDeletePod(deletePodCount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: err.Error(),
			Type:    1,
		}
		return
	}

	var podName string
	for _, podId := range deletePodCount.Pods {
		podName = fmt.Sprintf("%s-%d", static.SENDER_POD_LABEL, podId)
		_, ok := static.SenderPodIPMap[podName]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pod id not available. invalid pod id."})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("pod %s-%d not available. invalid pod id.", static.SENDER_POD_LABEL, podId),
				Type:    1,
			}
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("pod %s-%d will be deleted", static.SENDER_POD_LABEL, podId)})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("pod %s-%d will be deleted", static.SENDER_POD_LABEL, podId),
				Type:    0,
			}
		}

	}

	clientset, err := utils.CreateK8sClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create k8s client" + err.Error()})
		static.EventChannel <- static.SseData{
			Message: "unable to create k8s client",
			Type:    1,
		}
		return
	}

	totalSenderPods, err := utils.GetTotalPodCountInStatefulSet(clientset, static.K8S_NAMESPACE, static.SENDER_POD_LABEL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to get total pod count",
		})
		return
	}

	if len(deletePodCount.Pods) > 0 {
		err = utils.DeletePodInStatefulSet(clientset, static.K8S_NAMESPACE, podName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete pods" + err.Error()})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("unable to delete pod %s", podName),
				Type:    1,
			}
			return
		}
	} else if len(deletePodCount.Pods) == 0 {
		c.JSON(http.StatusOK, gin.H{"msg": "all pods will be deleted"})
		static.EventChannel <- static.SseData{
			Message: "all pods will be deleted",
			Type:    0,
		}
		for podId := range totalSenderPods {
			podName = fmt.Sprintf("%s-%d", static.SENDER_POD_LABEL, podId)
			err = utils.DeletePodInStatefulSet(clientset, static.K8S_NAMESPACE, podName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete pods" + err.Error()})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("unable to delete pod %s", podName),
					Type:    1,
				}
				return
			}
		}
	}

	clear(static.SenderPodIPMap)

	if err := utils.CreatePodMap(int(totalSenderPods)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create pod map",
		})
		return
	}

	fmt.Println("static.ConfigData.TraderCount=", static.ConfigData.TraderCount)

	static.EventChannel <- static.SseData{
		Message: "pod deleted",
		Type:    0,
	}

}
