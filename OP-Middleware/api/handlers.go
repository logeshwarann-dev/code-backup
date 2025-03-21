package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"net/http"
	static "op-middleware/static"
	"op-middleware/utils"
	"os"
	"path/filepath"
	"strings"

	// logger "op-middleware/logging"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

func HandleOPSRate(c *gin.Context) {
	var opsRateData static.OPSRate
	if static.ConfigData.TraderCount == 0 { // check for ongoing process before changing ops
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}
	if err := c.ShouldBindJSON(&opsRateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Invalid request body" + err.Error(),
			Type:    1,
		}
		return
	}

	if err := utils.ValidateChangeOPS(opsRateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "validation error: " + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: err.Error(),
			Type:    1,
		}
		return
	}

	if opsRateData.Type == 0 {
		for _, pod := range opsRateData.Pods {
			if _, ok := static.SenderPodIPMap[static.SENDER_POD_LABEL+"-"+strconv.Itoa(pod)]; !ok {
				c.JSON(http.StatusBadRequest, gin.H{
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
				c.JSON(http.StatusBadRequest, gin.H{
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

	c.JSON(http.StatusOK, gin.H{
		"message": "New OPS Received received successfully.",
	})

	if err := utils.SetOPSRateInPods(opsRateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error in changing ops rate: %v", err.Error()),
		})
	}

	static.EventChannel <- static.SseData{
		Message: "Change OPS Completed",
		Type:    0,
	}
}

func UpdateRecords(c *gin.Context) {

	if static.ConfigData.TraderCount == 0 { // check for ongoing process before changing ops
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}

	var updateRecordsData static.RecordsArray
	if err := c.ShouldBindJSON(&updateRecordsData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Invalid request body" + err.Error(),
			Type:    1,
		}
		return
	}

	for i, record := range updateRecordsData.Records {
		if err := utils.ValidateRecord(record); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation error" + strconv.Itoa(i) + err.Error(),
			})
			static.EventChannel <- static.SseData{
				Message: "invalid requet: " + err.Error(),
				Type:    1,
			}
			return
		}
	}

	fmt.Println("Records received successfully")
	c.JSON(http.StatusOK, gin.H{
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

		message := static.RecordsArray{
			Records: updateRecordsData.Records,
		}

		reqMsg, err := json.Marshal(message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
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

		message := static.RecordsArray{
			Records: updateRecordsData.Records,
		}

		reqMsg, err := json.Marshal(message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
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

func DeleteRecords(c *gin.Context) {

	if static.ConfigData.TraderCount == 0 { // check for ongoing process before changing ops
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}

	var data static.DeleteRecordStruct
	if err := c.ShouldBindBodyWithJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request format is not proper"})
		static.EventChannel <- static.SseData{
			Message: "Request format is not proper",
			Type:    1,
		}
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
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "No Instrument Id present as " + strconv.Itoa(data.InstId),
		})
		static.EventChannel <- static.SseData{
			Message: "Instrument Id not available: " + strconv.Itoa(data.InstId),
			Type:    1,
		}
		return
	}

	fmt.Println("Record delete request received successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Record will be deleted.",
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	for podName, podIP := range static.SenderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/deleteRecords"

		message := static.DeleteRecordStruct{
			InstId: data.InstId,
		}

		reqMsg, err := json.Marshal(message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
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

		message := static.DeleteRecordStruct{
			InstId: data.InstId,
		}

		reqMsg, err := json.Marshal(message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
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

func DeleteOrders(c *gin.Context) {

	if static.ConfigData.TraderCount == 0 { // check for ongoing process before changing ops
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}

	var deleteOrder static.DeleteOrder

	if err := c.ShouldBindJSON(&deleteOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Invalid request:" + err.Error(),
			Type:    1,
		}
		return
	}

	if err := utils.DeleteOrdersValidationCheck(deleteOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation error" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Invalid request:" + err.Error(),
			Type:    1,
		}
		return
	}

	recordCheck := 0
	for _, record := range static.RECORDS {
		if record.InstrumentID == deleteOrder.InstrumentID {
			recordCheck = 1
			if record.Product_ID != deleteOrder.ProductID {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Invalid Product Id: " + strconv.Itoa(deleteOrder.ProductID),
				})
				static.EventChannel <- static.SseData{
					Message: "Invalid Product Id: " + strconv.Itoa(deleteOrder.ProductID),
					Type:    1,
				}
				return
			}
			break
		}
	}

	if recordCheck == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "No Instrument Id present as " + strconv.Itoa(deleteOrder.InstrumentID),
		})
		static.EventChannel <- static.SseData{
			Message: "Instrument Id not available: " + strconv.Itoa(deleteOrder.InstrumentID),
			Type:    1,
		}
		return
	}

	fmt.Println("Orders delete request received successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Orders delete request received successfully",
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	// Iterate over pods
	for podName, podIP := range static.SenderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/deleteOrders"

		message := static.DeleteOrder{
			InstrumentID: deleteOrder.InstrumentID,
			ProductID:    deleteOrder.ProductID,
			Type:         deleteOrder.Type,
			Ids:          deleteOrder.Ids,
		}

		reqMsg, err := json.Marshal(message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
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

		message := static.DeleteOrder{
			InstrumentID: deleteOrder.InstrumentID,
			ProductID:    deleteOrder.ProductID,
		}

		reqMsg, err := json.Marshal(message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
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

func GetRecords(c *gin.Context) {
	fmt.Println("Records sent successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Records Sent successfully",
		"records": static.RECORDS,
	})

}

func SetConfig(c *gin.Context) {

	// static.ConfigData.UniqueClientIdentifier = true

	if static.ConfigData.TraderCount != 0 { // check for ongoing process before applying configuration
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping is in progress. shutdown the process to apply new configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping is in progress. shutdown the process to apply new configuration",
			Type:    1,
		}

		return
	}

	if !static.RECORDS_FILE_UPLOADED { // check for ongoing process before applying configuration
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Records file not uploaded",
		})
		static.EventChannel <- static.SseData{
			Message: "Records file not uploaded",
			Type:    1,
		}

		return
	}

	if err := c.ShouldBindJSON(&static.ConfigData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "config static.ConfigData not in proper format" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "configuation values are incorrect" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		return
	}

	if err := utils.ValidateConfig(static.ConfigData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		return
	}

	fmt.Println("Config received: ", static.ConfigData)
	static.EventChannel <- static.SseData{
		Message: "Config received. Processing",
		Type:    0,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config applied successfully",
	})

	static.TARGETENV = static.ConfigData.TargetEnv
	static.ConfigData.ProcessType = 0
	fmt.Println("Log Flag: ", static.ConfigData.LogFlag)

	clientset, err := utils.GetK8sClientset()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create k8s client" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "unable to create k8s client" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		return
	}

	senderTargetCount := int(math.Ceil(float64(static.ConfigData.TraderCount) / float64(static.SENDER_CONNECTION_COUNT)))

	err = utils.ScaleStatefulSet(clientset, static.K8S_NAMESPACE, static.SENDER_POD_LABEL, senderTargetCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while generating sender pods" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Error while generating sender pods" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		return
	}

	err = utils.ScaleStatefulSet(clientset, static.K8S_NAMESPACE, static.TRADER_POD_LABEL, static.TRADER_POD_COUNT)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while generating trader pods" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Error while generating trader pods" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		return
	}

	if err := utils.CreatePodMap(senderTargetCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create pod map" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "unable to create pod map" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		return
	}

	responses := make([]map[string]interface{}, 0)

	if err := utils.SetConfigInTraderPods(static.ConfigData, static.TraderPodIPMap, responses); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error setting config in trader pods:" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Error setting config in trader pods:" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		clear(static.TraderPodIPMap)
		return
	}

	static.CalculatedSenderThrottle = static.ConfigData.ThrottleValue / (len(static.SenderPodIPMap) * static.TRADER_CONNECTION_COUNT)

	if err := utils.SetConfigInSenderPods(static.ConfigData, static.SenderPodIPMap, responses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error setting config in sender pods:" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Error setting config in sender pods:" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		clear(static.SenderPodIPMap)
		clear(static.TraderPodIPMap)
	}

	records, err := utils.ReadRecords(static.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error reading records" + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Error reading records" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		clear(static.SenderPodIPMap)
		clear(static.TraderPodIPMap)
		return
	}

	static.RECORDS = append(static.RECORDS, records...)

	if err := utils.SetRecordsInPods(static.SenderPodIPMap, static.RECORDS, responses); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error setting records in sender pods: " + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Error setting records in sender pods:" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		clear(static.SenderPodIPMap)
		clear(static.TraderPodIPMap)
		clear(static.RECORDS)
		return
	}

	if err := utils.SetRecordsInPods(static.TraderPodIPMap, static.RECORDS, responses); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error setting records in trader pods: " + err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Error setting records in trader pods:" + err.Error(),
			Type:    1,
		}
		static.ConfigData = static.Config{} // reset config
		clear(static.SenderPodIPMap)
		clear(static.TraderPodIPMap)
		clear(static.RECORDS)
		return
	}

	fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

	static.PROCESS_STATE = 1
	static.EventChannel <- static.SseData{
		Message: "Set Config Completed",
		Type:    0,
	}
	c.JSON(http.StatusOK, gin.H{
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

	if static.ConfigData.TraderCount == 0 { // check for ongoing process before generating pattern
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}

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
		static.EventChannel <- static.SseData{
			Message: "Peak Generation in progress",
			Type:    0,
		}
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

		if generatePatternData.ProcessType == static.SYNC_ORDER_PUMPING_TYPE {

			if len(static.SyncPeakPodIPMap) >= static.SYNC_PEAK_POD_LIMIT {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("sync peak pod limit reached. current pod count: %v", len(static.SyncPeakPodIPMap)),
				})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("sync peak pod limit reached. current pod count: %v", len(static.SyncPeakPodIPMap)),
					Type:    1,
				}
				return
			}
			static.EventChannel <- static.SseData{
				Message: "Sync Peak Generation initiated",
				Type:    0,
			}

			totalSenderPods := len(opsRateData.Pods) - 1 // 1 pod will act as peak generator
			static.CalculatedSenderThrottle = baseOps / (totalSenderPods * static.SENDER_CONNECTION_COUNT)
			opsRateData.Throttle = static.CalculatedSenderThrottle
			opsRateData.Type = 0 //order pumping type - trader or sender

			if len(generatePatternData.Pods) == 0 {
				for peakPodMember, peakPodName := range static.MemberPodMap {
					if strings.Contains(peakPodName, static.TRADER_POD_LABEL) {
						continue
					}
					static.SyncPeakPodIPMap[peakPodName] = static.SenderPodIPMap[peakPodName]
					podId, _ := utils.ExtractPodId(peakPodName)
					generatePatternData.Pods = append(generatePatternData.Pods, podId)
					fmt.Println("In Peak Generation,", "| PeakPodIPMap: ", static.SyncPeakPodIPMap, "| Member: ", peakPodMember, "| Sender OPS RATE: ", opsRateData)
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

			c.JSON(http.StatusOK, gin.H{"message": "Peak Generation completed"})

			static.EventChannel <- static.SseData{
				Message: "Peak Generation completed",
				Type:    0,
			}

			fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

			return

		} else if generatePatternData.ProcessType == static.ASYNC_ORDER_PUMPING_OEOMOC_TYPE {

			if len(static.AsyncEMCPeakPodIPMap) >= static.ASYNC_EMC_PEAK_POD_LIMIT {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("async oe-om-oc pod limit reached. current pod count: %v", len(static.AsyncEMCPeakPodIPMap)),
				})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("async oe-om-oc pod limit reached. current pod count: %v", len(static.AsyncEMCPeakPodIPMap)),
					Type:    1,
				}
				return
			}

			static.EventChannel <- static.SseData{
				Message: "Async OE_OM_OC Peak Generation initiated",
				Type:    0,
			}

			totalSenderPods := len(opsRateData.Pods) // all existing sender pods will get baseOPS
			static.CalculatedSenderThrottle = baseOps / (totalSenderPods * static.SENDER_CONNECTION_COUNT)
			opsRateData.Throttle = static.CalculatedSenderThrottle
			opsRateData.Type = 0 //order pumping type - trader or sender

			// setting base ops in sender pods
			if err := utils.SetOPSRateInPods(opsRateData); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("error in changing ops rate: %v", err.Error()),
				})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("error in changing ops rate: %v", err.Error()),
					Type:    1,
				}
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

			fmt.Println("In EMC Peak Gen, pattern data: ", generatePatternData)

			newPodIPMap, err := utils.AddPodsWrapper(1)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("error in creating peak pod: %v", err.Error()),
				})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("error in creating peak pod: %v", err.Error()),
					Type:    1,
				}
				return
			}
			maps.Copy(static.AsyncEMCPeakPodIPMap, newPodIPMap)
			// fmt.Println("AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap)
			var responses []map[string]interface{}

			for podName := range static.AsyncEMCPeakPodIPMap { // restart the pod
				if err := utils.RestartPod(podName); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": fmt.Sprintf("error while restarting peak pod: %v", err.Error()),
					})
					static.EventChannel <- static.SseData{
						Message: fmt.Sprintf("error while restarting peak pod: %v", err.Error()),
						Type:    1,
					}
					return
				}
			}
			static.ConfigData.ProcessType = 1
			static.ConfigData.AppSystemVendorName = generatePatternData.AppSystemVendorName
			static.ConfigData.AppSystemVersion = generatePatternData.AppSystemVersion

			if err := utils.SetConfigInSenderPods(static.ConfigData, static.AsyncEMCPeakPodIPMap, responses); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error setting config in peak pods: %v", err.Error())})
				return
			}

			if err := utils.SetRecordsInPods(static.AsyncEMCPeakPodIPMap, static.RECORDS, responses); err != nil {
				fmt.Printf("Error setting records in sender pods: %v", err.Error())
			}

			fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

			c.JSON(http.StatusOK, gin.H{"message": "Peak Generation completed"})

			static.EventChannel <- static.SseData{
				Message: "Peak Generation completed",
				Type:    0,
			}

			return

		} else if generatePatternData.ProcessType == static.ASYNC_ORDER_PUMPING_IOC_TYPE {

			if len(static.AsyncIOCPeakPodIPMap) >= static.ASYNC_IOC_PEAK_POD_LIMIT {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("async ioc pod limit reached. current pod count: %v", len(static.AsyncIOCPeakPodIPMap)),
				})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("async ioc pod limit reached. current pod count: %v", len(static.AsyncIOCPeakPodIPMap)),
					Type:    1,
				}
				return
			}

			static.EventChannel <- static.SseData{
				Message: "Async IOC Peak Generation initiated",
				Type:    0,
			}
			totalSenderPods := len(opsRateData.Pods) // all existing sender pods will get baseOPS
			static.CalculatedSenderThrottle = baseOps / (totalSenderPods * static.SENDER_CONNECTION_COUNT)
			opsRateData.Throttle = static.CalculatedSenderThrottle
			opsRateData.Type = 0 //order pumping type - trader or sender

			// setting base ops in sender pods
			if err := utils.SetOPSRateInPods(opsRateData); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("error in changing ops rate: %v", err.Error()),
				})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("error in changing ops rate: %v", err.Error()),
					Type:    1,
				}
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

			fmt.Println("In IOC Peak Gen, pattern data: ", generatePatternData)

			newPodIPMap, err := utils.AddPodsWrapper(1)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("error in creating peak pod: %v", err.Error()),
				})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("error in creating peak pod: %v", err.Error()),
					Type:    1,
				}
				return
			}
			maps.Copy(static.AsyncIOCPeakPodIPMap, newPodIPMap)
			// fmt.Println("AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap)
			var responses []map[string]interface{}

			for podName := range static.AsyncIOCPeakPodIPMap { // restart the pod
				if err := utils.RestartPod(podName); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": fmt.Sprintf("error while restarting peak pod: %v", err.Error()),
					})
					static.EventChannel <- static.SseData{
						Message: fmt.Sprintf("error while restarting peak pod: %v", err.Error()),
						Type:    1,
					}
					return
				}
			}
			static.ConfigData.ProcessType = 2

			if err := utils.SetConfigInSenderPods(static.ConfigData, static.AsyncIOCPeakPodIPMap, responses); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error setting config in peak pods: %v", err.Error())})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("error setting config in peak pods: %v", err.Error()),
					Type:    1,
				}
				return
			}

			if err := utils.SetRecordsInPods(static.AsyncIOCPeakPodIPMap, static.RECORDS, responses); err != nil {

				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("error setting records in sender pods: %v", err.Error())})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("error setting records in sender pods: %v", err.Error()),
					Type:    1,
				}
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Peak Generation completed"})

			static.EventChannel <- static.SseData{
				Message: "Peak Generation completed",
				Type:    0,
			}
			fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

			return
		}

	}

	if len(generatePatternData.Pods) == 0 {
		generatePatternData.Pods = utils.GetActivePodIds("mock-sender-")
	}

	fmt.Println("In Pattern gen, sender pod data: ", generatePatternData)

	if err := utils.SendPatternGenerateDataToPods(generatePatternData, static.SENDER_POD_LABEL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error in sending Pattern to sender pods: %v", err.Error()),
		})
		static.EventChannel <- static.SseData{
			Message: fmt.Sprintf("error in sending Pattern to sender pods: %v", err.Error()),
			Type:    1,
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pattern is set in pods"})

	static.EventChannel <- static.SseData{
		Message: "Pattern is set in pods",
		Type:    0,
	}

}

func AddSlidingPrice(c *gin.Context) {

	if static.ConfigData.TraderCount == 0 { // check for ongoing process before changing ops
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}

	var slidingPriceData static.PriceRangeChangeDetails

	if err := c.ShouldBindJSON(&slidingPriceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Invalid request:" + err.Error(),
			Type:    1,
		}
		return
	}

	if err := utils.ValidatePriceRangeChange(slidingPriceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Invalid request:" + err.Error(),
			Type:    1,
		}
		return
	}

	var totalRecordsFound []int
	for _, record := range static.RECORDS {
		_, ok := slidingPriceData.Instruments[strconv.Itoa(record.InstrumentID)]
		if ok {
			totalRecordsFound = append(totalRecordsFound, 1)
		}
	}

	if len(totalRecordsFound) != len(slidingPriceData.Instruments) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("One or many Instrument Ids not available: only %v records found.", len(totalRecordsFound)),
		})
		static.EventChannel <- static.SseData{
			Message: fmt.Sprintf("One or many Instrument Ids not available: only %v records found.", len(totalRecordsFound)),
			Type:    1,
		}
		return
	}

	if err := utils.ValidatePrices(slidingPriceData.Instruments); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "Invalid request:" + err.Error(),
			Type:    1,
		}
		return
	}

	if err := utils.SetSlidingPriceInPods(slidingPriceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "error in sliding price",
			"details": err.Error(),
		})
		static.EventChannel <- static.SseData{
			Message: "error in sliding price:" + err.Error(),
			Type:    1,
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sliding Price is set in pods"})

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
	if static.ConfigData.TraderCount == 0 { // check for ongoing process before adding pods
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}

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
	c.JSON(http.StatusOK, gin.H{"message": "new pods will be created"})

	static.EventChannel <- static.SseData{
		Message: "pod creation in progress",
		Type:    0,
	}
	// Add Trader Pod
	if addPodPayload.Count == static.TRADER_POD_MAGIC_NUM {

		traderPodName := static.TRADER_POD_LABEL + "-0"
		podName, ok := static.MemberPodMap[static.TRADER_MEMBERID]
		if ok {
			if podName == traderPodName {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "trader pod already exists.",
				})
				static.EventChannel <- static.SseData{
					Message: "trader pod already exists.",
					Type:    1,
				}
				return
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": podName + " pod uses 1046 member. delete that pod to create trader",
				})
				static.EventChannel <- static.SseData{
					Message: podName + " pod uses 1046 member. delete that pod to create trader",
					Type:    1,
				}
				return
			}
		}

		availableMemberCount := utils.CheckAvailableMemberIds()
		if availableMemberCount == 0 {
			static.EventChannel <- static.SseData{
				Message: "no session ids left. cannot create pods",
				Type:    1,
			}
			return
		}
		addPodPayload.Count = 1
		responses := make([]map[string]interface{}, 0)
		if err := utils.SetConfigInTraderPods(static.ConfigData, static.TraderPodIPMap, responses); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error setting config in trader pods:" + err.Error(),
			})
			static.EventChannel <- static.SseData{
				Message: "Error setting config in trader pods:" + err.Error(),
				Type:    1,
			}
			static.ConfigData = static.Config{} // reset config
			clear(static.TraderPodIPMap)
			return
		}

		if err := utils.SetRecordsInPods(static.TraderPodIPMap, static.RECORDS, responses); err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error setting records in trader pods: " + err.Error(),
			})
			static.EventChannel <- static.SseData{
				Message: "Error setting records in trader pods:" + err.Error(),
				Type:    1,
			}
			static.ConfigData = static.Config{} // reset config
			clear(static.SenderPodIPMap)
			clear(static.TraderPodIPMap)
			clear(static.RECORDS)
			return
		}

		fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

		static.EventChannel <- static.SseData{
			Message: "1 trader pod created",
			Type:    0,
		}
		return
	}

	// Add Sender Pod
	reqPodCount := addPodPayload.Count
	podsStatusFlag, inactivePodsArr := utils.GetPumpingStatusOfPods()

	availableMemberCount := utils.CheckAvailableMemberIds()
	fmt.Println("Available Member IDs: ", availableMemberCount)
	if availableMemberCount == 0 {
		// c.JSON(http.StatusForbidden, gin.H{"error": "no session ids left. cannot create pods"})
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
		existingPodStatus, totalPodsStarted, _ := utils.UseExistingPodsInStatefulSet(reqPodCount)

		if existingPodStatus && totalPodsStarted == reqPodCount {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d pod(s) activated", reqPodCount)})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("%d pod(s) created", reqPodCount),
				Type:    0,
			}
			return
		} else if existingPodStatus && totalPodsStarted < reqPodCount {
			reqPodCount -= totalPodsStarted
		}
	}

	_, err := utils.UseNewPodsInStatefulSet(reqPodCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("error creating new %v pod: %v", reqPodCount, err.Error())})
		static.EventChannel <- static.SseData{
			Message: fmt.Sprintf("error creating new %v pod: %v", reqPodCount, err.Error()),
			Type:    1,
		}
		return
	}

	// fmt.Println("Member Pod Map: ", static.MemberPodMap)
	// c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d pods are created", reqPodCount)})

	fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

	static.EventChannel <- static.SseData{
		Message: fmt.Sprintf("%d pod(s) created", reqPodCount),
		Type:    0,
	}

}

func DeletePods(c *gin.Context) {

	var deletePodCount static.DeletePodCount

	if static.ConfigData.TraderCount == 0 { // check for ongoing process before deleting pods
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping not started. apply configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping not started. apply configuration",
			Type:    1,
		}
		return
	}

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

	static.EventChannel <- static.SseData{
		Message: "pod deletion in progress",
		Type:    0,
	}

	var podName string
	var podId int
	for _, podId = range deletePodCount.Pods {
		podName = fmt.Sprintf("%s-%d", static.SENDER_POD_LABEL, podId)
		// _, senderPodIdCheck := static.SenderPodIPMap[podName]
		// _, syncPeakPodIdCheck := static.SyncPeakPodIPMap[podName]
		// _, asyncEMCPodIdCheck := static.AsyncEMCPeakPodIPMap[podName]
		// _, asyncIOCPodIdCheck := static.AsyncIOCPeakPodIPMap[podName]

		if podId == static.TRADER_POD_MAGIC_NUM {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("trader pod %s-0 will be deleted", static.TRADER_POD_LABEL)})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("trader pod %s-0 will be deleted", static.TRADER_POD_LABEL),
				Type:    0,
			}
			break
		}

		isPodPresent := utils.CheckPodAvailability(podId)

		if !isPodPresent {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pod id not available. invalid pod id."})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("pod %s-%d not available. invalid pod id.", static.SENDER_POD_LABEL, podId),
				Type:    1,
			}
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("pod %s-%d will be deleted", static.SENDER_POD_LABEL, podId)})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("pod %s-%d will be deleted", static.SENDER_POD_LABEL, podId),
				Type:    0,
			}
		}

	}

	clientset, err := utils.GetK8sClientset()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		static.EventChannel <- static.SseData{
			Message: err.Error(),
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

		if podId == static.TRADER_POD_MAGIC_NUM {
			podName = static.TRADER_POD_LABEL + "-0"
			if err = utils.DeletePodInStatefulSet(clientset, static.K8S_NAMESPACE, podName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete pods" + err.Error()})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("unable to delete pod %s: %v", podName, err.Error()),
					Type:    1,
				}
				return
			}
			clear(static.TraderPodIPMap)
		} else {
			if err = utils.DeletePodInStatefulSet(clientset, static.K8S_NAMESPACE, podName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete pods" + err.Error()})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("unable to delete pod %s: %v", podName, err.Error()),
					Type:    1,
				}
				return
			}
			static.ConfigData.TraderCount -= static.TRADER_CONNECTION_COUNT
		}

	} else if len(deletePodCount.Pods) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "all pods will be deleted"})
		static.EventChannel <- static.SseData{
			Message: "all pods will be deleted",
			Type:    0,
		}
		for podId := range totalSenderPods {
			podName = fmt.Sprintf("%s-%d", static.SENDER_POD_LABEL, podId)
			if err = utils.DeletePodInStatefulSet(clientset, static.K8S_NAMESPACE, podName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete pods" + err.Error()})
				static.EventChannel <- static.SseData{
					Message: fmt.Sprintf("unable to delete pod %s: %v", podName, err.Error()),
					Type:    1,
				}
				return
			}
		}
		static.ConfigData.TraderCount = 0
		static.PROCESS_STATE = 0
		static.RECORDS_FILE_UPLOADED = false
	}

	_, ok := static.SenderPodIPMap[podName]
	if ok {
		clear(static.SenderPodIPMap)
		// delete(static.SenderPodIPMap, podName)

		if err := utils.CreatePodMap(int(totalSenderPods)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "unable to create pod map",
			})
			static.EventChannel <- static.SseData{
				Message: fmt.Sprintf("unable to unable to create pod map: %v", err.Error()),
				Type:    1,
			}
			return
		}
	}

	if len(deletePodCount.Pods) == 0 {
		clear(static.MemberPodMap)
		clear(static.SenderPodIPMap)
		clear(static.TraderPodIPMap)
		clear(static.AsyncEMCPeakPodIPMap)
		clear(static.AsyncIOCPeakPodIPMap)
		clear(static.SyncPeakPodIPMap)
		static.RECORDS = nil
	}
	_, ok = static.SyncPeakPodIPMap[podName]
	if ok {
		delete(static.SyncPeakPodIPMap, podName)
	}
	_, ok = static.AsyncEMCPeakPodIPMap[podName]
	if ok {
		delete(static.AsyncEMCPeakPodIPMap, podName)
	}
	_, ok = static.AsyncIOCPeakPodIPMap[podName]
	if ok {
		delete(static.AsyncIOCPeakPodIPMap, podName)
	}

	fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

	fmt.Println("static.ConfigData.TraderCount=", static.ConfigData.TraderCount)

	static.EventChannel <- static.SseData{
		Message: "pod deletion completed",
		Type:    0,
	}

}

func ShutDownPods(c *gin.Context) {

	clientset, err := utils.GetK8sClientset()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		static.EventChannel <- static.SseData{
			Message: err.Error(),
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

	if totalSenderPods == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no pods are running"})
		static.EventChannel <- static.SseData{
			Message: "no pods are running",
			Type:    1,
		}
		return
	}

	clear(static.MemberPodMap)
	clear(static.SenderPodIPMap)
	clear(static.TraderPodIPMap)
	clear(static.AsyncEMCPeakPodIPMap)
	clear(static.AsyncIOCPeakPodIPMap)
	clear(static.SyncPeakPodIPMap)
	static.RECORDS = nil
	senderPodCount := len(static.SenderPodIPMap)
	traderPodCount := len(static.TraderPodIPMap)

	static.EventChannel <- static.SseData{
		Message: "pod cleanup in progress",
		Type:    0,
	}

	c.JSON(http.StatusOK, gin.H{"message": "all pods will be removed"})

	err = utils.ScaleStatefulSet(clientset, static.K8S_NAMESPACE, static.SENDER_POD_LABEL, senderPodCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while deleting sender pods" + err.Error(),
		})
		fmt.Println("Error while deleting sender pods" + err.Error())
		static.EventChannel <- static.SseData{
			Message: "Error while deleting sender pods" + err.Error(),
			Type:    1,
		}
		return
	}

	static.EventChannel <- static.SseData{
		Message: "sender pods are removed",
		Type:    0,
	}

	err = utils.ScaleStatefulSet(clientset, static.K8S_NAMESPACE, static.TRADER_POD_LABEL, traderPodCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while deleting trader pods" + err.Error(),
		})
		fmt.Println("Error while deleting trader pods" + err.Error())
		static.EventChannel <- static.SseData{
			Message: "Error while deleting trader pods" + err.Error(),
			Type:    1,
		}
		return
	}

	static.ConfigData.TraderCount = 0
	static.PROCESS_STATE = 0
	static.RECORDS_FILE_UPLOADED = false

	static.EventChannel <- static.SseData{
		Message: "trader pod is removed",
		Type:    0,
	}

	fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

	fmt.Println("static.ConfigData.TraderCount=", static.ConfigData.TraderCount)

}

func StoreRecordsFile(c *gin.Context) {

	if static.ConfigData.TraderCount != 0 { // check for ongoing process before applying configuration
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "order pumping is in progress. shutdown the process to apply new configuration",
		})
		static.EventChannel <- static.SseData{
			Message: "order pumping is in progress. shutdown the process to apply new configuration",
			Type:    1,
		}
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file" + err.Error()})
		static.EventChannel <- static.SseData{
			Message: "Failed to retrieve file" + err.Error(),
			Type:    1,
		}
		return
	}
	fmt.Println("File received!")

	savePath := "Records/mockRecords.txt"

	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory" + err.Error()})
		static.EventChannel <- static.SseData{
			Message: "Failed to create directory" + err.Error(),
			Type:    1,
		}
		return
	}

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file" + err.Error()})
		static.EventChannel <- static.SseData{
			Message: "Failed to save file" + err.Error(),
			Type:    1,
		}
		return
	}

	fmt.Println("File saved Successfully!")

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("File '%s' uploaded successfully!", file.Filename),
		"path":    savePath,
	})

	static.EventChannel <- static.SseData{
		Message: fmt.Sprintf("File '%s' uploaded!", file.Filename),
		Type:    0,
	}

	static.RECORDS_FILE_UPLOADED = true
}
