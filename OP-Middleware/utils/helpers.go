package utils

import (
	// logger "op-middleware/logging"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	static "op-middleware/static"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var K8sClientSet *kubernetes.Clientset

func ReadRecords(filename string) ([]static.Record, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []static.Record
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entries := strings.Split(line, ",")

		if len(entries) < 9 {
			fmt.Println("[ERR] Invalid line:", line)
			continue
		}

		instrumentID, err0 := strconv.Atoi(entries[0])
		lowerLimit, err1 := strconv.Atoi(entries[5])
		upperLimit, err2 := strconv.Atoi(entries[6])
		bidIntrvl, err3 := strconv.Atoi(entries[7])
		maxTrdQty, err4 := strconv.Atoi(entries[8])
		minLot, err5 := strconv.Atoi(entries[9])
		product_id, err6 := strconv.Atoi(entries[10])
		partition_id, err7 := strconv.Atoi(entries[11])

		fmt.Printf("[INFO] Instrument ID: %v | Bid Interval: %v | MaxTrdQty: %v |  MinLot value: %v | Product ID: %v | Partition ID : %v\n", instrumentID, bidIntrvl, maxTrdQty, minLot, product_id, partition_id)

		if err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
			fmt.Println("[ERR] Error parsing limits:", err0, err1, err2, err3, err4, err5, err6, err7)
			continue
		}

		records = append(records, static.Record{
			InstrumentID: instrumentID,
			LowerLimit:   lowerLimit,
			UpperLimit:   upperLimit,
			BidInterval:  bidIntrvl,
			MaxTrdQty:    maxTrdQty,
			MinLot:       minLot,
			Product_ID:   product_id,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil

}

func ScaleStatefulSet(clientset *kubernetes.Clientset, namespace, statefulSetName string, replicas int) error {
	// Get the StatefulSet
	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Update the replica count
	newReplicaCount := int32(replicas)
	statefulSet.Spec.Replicas = &newReplicaCount

	// Update the StatefulSet
	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Scaled StatefulSet %s to %d replicas\n", statefulSetName, replicas)
	return nil
}

func GetPumpingStatusOfPods() (bool, []string) {
	reqBody := `{"message": "Check Pumping status"}`
	podResponses := FanOutRequestsToPods("/api/v1/orderPumpingStatus", reqBody, static.SENDER_POD_LABEL, "8080")
	var inactivePods []string
	fmt.Println("Pumping status Response from all pods: ", podResponses)
	for _, response := range podResponses {
		jsonData, _ := json.Marshal(response)
		var eachPodResponse static.PodResponse
		var podStatus static.PodActiveResponse
		if err := json.Unmarshal(jsonData, &eachPodResponse); err != nil {
			fmt.Println("error in unmarshalling pod response: ", err.Error())
			return false, nil
		}
		err := json.Unmarshal([]byte(eachPodResponse.Response), &podStatus)
		if err != nil {
			fmt.Printf("Failed to unmarshal JSON: %v", err)
			return false, nil
		}

		fmt.Println("Pumping status Response from pod: ", podStatus)
		if !podStatus.Status {
			fmt.Println("Inactive Pod: ", podStatus.PodName)
			inactivePods = append(inactivePods, podStatus.PodName)
		}
	}

	if len(inactivePods) == 0 {
		fmt.Println("No inactive Pods.")
		return false, nil
	}
	return true, inactivePods
}

func UseExistingPodsInStatefulSet(desiredPodCount int) (bool, int, map[string]string) {

	pumpStatus, inactivePods := GetPumpingStatusOfPods()

	if !pumpStatus {
		return false, 0, nil
	}

	inactivePodsMap := make(map[string]string)

	for _, podname := range inactivePods {
		inactivePodsMap[podname] = static.SenderPodIPMap[podname]

		if len(inactivePodsMap) == desiredPodCount {
			break
		}
	}

	fmt.Println("Inactive Pod Map: ", inactivePodsMap)

	responses := make([]map[string]interface{}, 0)
	if err := SetConfigInSenderPods(static.ConfigData, inactivePodsMap, responses); err != nil {
		fmt.Printf("Error setting config in sender pods: %v", err.Error())
		static.EventChannel <- static.SseData{
			Message: fmt.Sprint("Error setting config in sender pods", err.Error()),
			Type:    1,
		}
	}

	if err := SetRecordsInPods(inactivePodsMap, static.RECORDS, responses); err != nil {
		fmt.Printf("Error setting records in sender pods: %v", err.Error())
	}

	static.ConfigData.TraderCount += (len(inactivePodsMap) * static.SENDER_CONNECTION_COUNT)

	return true, len(inactivePodsMap), inactivePodsMap

}

func UseNewPodsInStatefulSet(desiredPodCount int) (map[string]string, error) {
	clientset, err := GetK8sClientset()
	if err != nil {
		return nil, err
	}

	totalSenderPods, err := CreateNewPodsToStatefulSet(clientset, static.K8S_NAMESPACE, static.SENDER_POD_LABEL, desiredPodCount)
	if err != nil {
		return nil, fmt.Errorf("unable to add new pods %v", err.Error())
	}

	if err := CreatePodMap(int(totalSenderPods)); err != nil {
		return nil, fmt.Errorf("unable to create pod map %v", err.Error())
	}

	newPodsMap := make(map[string]string)

	newPodIndex := len(static.SenderPodIPMap) - desiredPodCount
	for range desiredPodCount {
		newPodKey := fmt.Sprintf("%s-%d", static.SENDER_POD_LABEL, newPodIndex)

		newPodsMap[newPodKey] = static.SenderPodIPMap[newPodKey]
		newPodIndex++
	}

	fmt.Println("New Pod Map: ", newPodsMap)

	responses := make([]map[string]interface{}, 0)
	static.ConfigData.TraderCount += (desiredPodCount * static.SENDER_CONNECTION_COUNT)
	if err := SetConfigInSenderPods(static.ConfigData, newPodsMap, responses); err != nil {
		return nil, fmt.Errorf("error setting config in sender pods: %v", err.Error())
	}

	if err := SetRecordsInPods(newPodsMap, static.RECORDS, responses); err != nil {
		return nil, fmt.Errorf("error setting records in sender pods: %v", err.Error())
	}

	fmt.Println("Total New Pods Created: ", desiredPodCount)

	fmt.Println("static.ConfigData.TraderCount=", static.ConfigData.TraderCount)

	return newPodsMap, nil
}

func CreateNewPodsToStatefulSet(clientset *kubernetes.Clientset, namespace, statefulSetName string, newPodCount int) (int32, error) {

	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	currentReplicas := *statefulSet.Spec.Replicas
	if currentReplicas == 12 {
		return 0, fmt.Errorf("cannot add new pod because already 12 pods are running")
	}

	newReplicas := currentReplicas + int32(newPodCount)
	statefulSet.Spec.Replicas = &newReplicas

	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return 0, err
	}

	fmt.Printf("%v pods are added to %v statefulset. Current Pods count: %v\n", newPodCount, statefulSetName, newReplicas)
	return newReplicas, nil
}

func CheckPodAvailability(podId int) bool {
	podName := fmt.Sprintf("%s-%d", static.SENDER_POD_LABEL, podId)
	_, senderPodIdCheck := static.SenderPodIPMap[podName]
	_, syncPeakPodIdCheck := static.SyncPeakPodIPMap[podName]
	_, asyncEMCPodIdCheck := static.AsyncEMCPeakPodIPMap[podName]
	_, asyncIOCPodIdCheck := static.AsyncIOCPeakPodIPMap[podName]
	podName = fmt.Sprintf("%s-%d", static.TRADER_POD_LABEL, podId)
	fmt.Println("Podname received: ", podName)
	_, traderPodIdCheck := static.TraderPodIPMap[podName]
	fmt.Println("Is Pod available: ", traderPodIdCheck)
	if !senderPodIdCheck && !syncPeakPodIdCheck && !asyncEMCPodIdCheck && !asyncIOCPodIdCheck && !traderPodIdCheck {
		return false
	}
	return true
}

func GetTotalPodCountInStatefulSet(clientset *kubernetes.Clientset, namespace, statefulSetName string) (int, error) {
	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}
	currentReplicaCount := int(*statefulSet.Spec.Replicas)
	GetPodHealthStatus(namespace, currentReplicaCount)
	return currentReplicaCount, nil
}

func ScaleDownReplicaInStatefulSet(clientset *kubernetes.Clientset, namespace, statefulSetName string, deletePodCount int) (int32, error) {

	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	currentReplicas := *statefulSet.Spec.Replicas
	if currentReplicas == 0 {
		return 0, fmt.Errorf("cannot delete pod because no pods are running")
	}

	// Scale up by reducing replicas
	newReplicas := currentReplicas - int32(deletePodCount)
	statefulSet.Spec.Replicas = &newReplicas

	// Update the StatefulSet
	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return 0, err
	}

	fmt.Printf("%v pods are removed in %v statefulset. Current Pods count: %v\n", deletePodCount, statefulSetName, newReplicas)
	return newReplicas, nil

}

func DeletePodInStatefulSet(clientset *kubernetes.Clientset, namespace, podName string) error {

	err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod %s: %v", podName, err)
	}
	fmt.Printf("Pod %s deleted successfully\n", podName)
	RemovePodInMemberMap(podName)
	return nil

}

func GetPodHealthStatus(appLabel string, podCount int) bool {
	reqBody := `{"message": "Health Check"}`
	responses := FanOutRequestsToPods("/api/v1/health-check", reqBody, appLabel, "8080")
	if len(responses) != podCount {
		fmt.Printf("All Pods are not healthy. Only %v pods are healthy and responses: %v\n", len(responses), responses)
		return false
	}
	for _, response := range responses {
		jsonData, _ := json.Marshal(response)
		var podHealthResponse static.PodResponse
		if err := json.Unmarshal(jsonData, &podHealthResponse); err != nil {
			fmt.Println("error in unmarshalling pod response: ", err.Error())
			return false
		}
		if !strings.Contains(podHealthResponse.Response, "is healthy") {
			fmt.Println("All Pods are not active. Received response as: ", podHealthResponse.Response)
			return false
		}
	}
	fmt.Println("All Pods are healthy.")
	return true
}

func FanOutRequestsToPods(endpoint string, reqMsg string, appLabel string, servicePort string) []map[string]interface{} {
	podsList, err := GetPodAddr(appLabel, static.K8S_NAMESPACE)
	if err != nil {
		fmt.Println("error in getting pod address: ", err.Error())
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	for _, pod := range podsList {
		podIP := pod.Status.PodIP
		podAddr := fmt.Sprintf("http://%s:%s%s", podIP, servicePort, endpoint)

		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

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

	wg.Wait()

	return responses
}

func GetPodAddr(appLabel string, namespace string) ([]v1.Pod, error) {

	clientset, err := GetK8sClientset()
	if err != nil {
		log.Fatal("unable to create k8s client")
		return nil, err
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=" + appLabel,
	})
	if err != nil {
		log.Fatalf("Error getting client pods: %v", err)
		return nil, err
	}

	allPods := pods.Items

	return allPods, nil
}

func CreateK8sClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
		return err
	}
	K8sClientSet = clientset
	return nil
}

func GetK8sClientset() (*kubernetes.Clientset, error) {
	if K8sClientSet != nil {
		return K8sClientSet, nil
	}
	return nil, fmt.Errorf("k8s client set is nil")
}

func CreatePodMap(targetCount int) error {
	for range 5 {
		time.Sleep(5 * time.Second)
		isPodActive := GetPodHealthStatus(static.SENDER_POD_LABEL, targetCount)
		if !isPodActive {
			continue
		}
		allPods, err := GetPodAddr(static.SENDER_POD_LABEL, static.K8S_NAMESPACE)
		if err != nil {
			fmt.Println("Error in getting pod address", err.Error())
			return err
		}

		for _, pod := range allPods {
			static.SenderPodIPMap[pod.Name] = pod.Status.PodIP
		}

		isPodActive = GetPodHealthStatus(static.TRADER_POD_LABEL, static.TRADER_POD_COUNT)
		if !isPodActive {
			continue
		}
		allPods, err = GetPodAddr(static.TRADER_POD_LABEL, static.K8S_NAMESPACE)
		if err != nil {
			fmt.Println("Error in getting pod address", err.Error())
			return err
		}
		for _, pod := range allPods {
			static.TraderPodIPMap[pod.Name] = pod.Status.PodIP
		}

		for podName := range static.SyncPeakPodIPMap {
			static.SyncPeakPodIPMap[podName] = static.SenderPodIPMap[podName]
		}
		for podName := range static.AsyncEMCPeakPodIPMap {
			static.AsyncEMCPeakPodIPMap[podName] = static.SenderPodIPMap[podName]
		}
		for podName := range static.AsyncIOCPeakPodIPMap {
			static.AsyncIOCPeakPodIPMap[podName] = static.SenderPodIPMap[podName]
		}

		if len(static.TraderPodIPMap) == static.TRADER_POD_COUNT && len(static.SenderPodIPMap) == targetCount {
			fmt.Println("POD MAP CREATED")
			break
		}

	}

	fmt.Println("SenderPodIPMap: ", static.SenderPodIPMap, "| SyncPeakPodIPMap: ", static.SyncPeakPodIPMap, "| AsyncEMCPeakPodIPMap: ", static.AsyncEMCPeakPodIPMap, "| AsyncIOCPeakPodIPMap: ", static.AsyncIOCPeakPodIPMap, "| MemberPodMap: ", static.MemberPodMap)

	return nil
}

func GetAllNodesInTheCluster(clientset *kubernetes.Clientset) ([]string, error) {
	nodesInCluster, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var nodesList []string
	for _, node := range nodesInCluster.Items {
		fmt.Println(" -", node.Name)
		nodesList = append(nodesList, node.Name)

	}

	return nodesList, nil
}

func SetRecordsInPods(podMap map[string]string, records []static.Record, responses []map[string]interface{}) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for podName, podIP := range podMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/setRecords"

		msg := static.RecordsArray{
			Records: records,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error : Error marshalling request - %v", err.Error())
		}

		wg.Add(1)
		go func(addr string, podName string) {
			defer wg.Done()
			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("POST", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in setting Records for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in setting Records for " + podName + " : " + err.Error())
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
						static.EventChannel <- static.SseData{Message: "Error in setting Records for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in setting Records for " + podName + " : " + err.Error())
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
					fmt.Println("[ERR] Error in setting Records in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in setting Records for " + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Records set successfully for " + podName + "and started order pumping", Type: 0}
				fmt.Println("Records set successfully for " + podName + "and started order pumping")
				mu.Unlock()
				break
			}
		}(podAddr, podName)

	}

	wg.Wait()

	return nil
}

func CheckAvailableMemberIds() int {
	if static.TARGETENV == 1 {
		return len(static.LabSessionIdSets) - len(static.MemberPodMap)
	} else if static.TARGETENV == 2 {
		return len(static.ProdSessionIdSets) - len(static.MemberPodMap)
	} else {
		return len(static.SimSessionIdSets) - len(static.MemberPodMap)
	}
}

func AssignMembersToPods(traders []static.Trader, sessionIdSets map[string][]static.Trader, podName string, specificKey string) []static.Trader {

	if len(specificKey) > 0 {
		static.MemberPodMap[specificKey] = podName
		traders = append(traders, sessionIdSets[specificKey]...)
		fmt.Println(specificKey, " Session IDs are assigned to Pod ", podName)
		return traders
	}

	for key, value := range sessionIdSets {
		if _, ok := static.MemberPodMap[key]; !ok {
			// if key == static.PEAK_POD_MEMBERID {
			// 	continue
			// }
			static.MemberPodMap[key] = podName
			traders = append(traders, value...)
			fmt.Println(key, " Session IDs are assigned to Pod ", podName)
			return traders
		}
	}
	fmt.Println("all Session IDs are already assigned to Pods")
	return []static.Trader{}
}

func RemovePodInMemberMap(podName string) {

	for key, value := range static.MemberPodMap {
		if value == podName {
			delete(static.MemberPodMap, key)
			fmt.Println(key, " and ", podName, " deleted in MemberPodMap")
		}
	}

}

func SetConfigInSenderPods(data static.Config, senderPodMap map[string]string, responses []map[string]interface{}) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for podName, podIP := range senderPodMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/setConfig"
		var traders []static.Trader

		if data.TargetEnv == 1 {
			traders = AssignMembersToPods(traders, static.LabSessionIdSets, podName, "")
		} else if data.TargetEnv == 2 {
			traders = AssignMembersToPods(traders, static.ProdSessionIdSets, podName, "")
		} else {
			traders = AssignMembersToPods(traders, static.SimSessionIdSets, podName, "")
		}

		fmt.Println("Member Pod Map: ", static.MemberPodMap)

		fmt.Println(podName, " Throttle value: ", static.CalculatedSenderThrottle)
		fmt.Println("UniqClientID: ", data.UniqueClientIdentifier)
		msg := static.Config{
			TotalOrderCount:        data.TotalOrderCount,
			TargetEnv:              data.TargetEnv,
			ModifyPercent:          data.ModifyPercent,
			CancelPercent:          data.CancelPercent,
			ThrottleValue:          static.CalculatedSenderThrottle,
			FileType:               data.FileType,
			TraderCount:            static.SENDER_CONNECTION_COUNT,
			HeartbeatValue:         data.HeartbeatValue,
			UniqueClientIdentifier: data.UniqueClientIdentifier,
			Traders:                traders,
			TradeFlag:              0,
			Duration:               data.Duration,
			ProcessType:            data.ProcessType,
			AppSystemVendorName:    data.AppSystemVendorName,
			AppSystemVersion:       data.AppSystemVersion,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error : Error marshalling request - %v", err.Error())
		}

		wg.Add(1)
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("POST", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in setting config for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in setting config for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Println("Error in http.client")
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in setting config for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in setting config for " + podName + " : " + err.Error())
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
					fmt.Println("[ERR] Error in setting config in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in setting config for" + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})

				fmt.Println("Config Set Response from Sender Pods: ", responses)

				static.EventChannel <- static.SseData{Message: "Config set successfully for " + podName, Type: 0}

				static.ConfigDetails.TotalOrderCount = data.TotalOrderCount
				static.ConfigDetails.TargetEnv = data.TargetEnv
				static.ConfigDetails.ModifyPercent = data.ModifyPercent
				static.ConfigDetails.CancelPercent = data.CancelPercent
				static.ConfigDetails.FileType = data.FileType
				static.ConfigDetails.TraderCount = data.TraderCount
				static.ConfigDetails.HeartbeatValue = data.HeartbeatValue
				static.ConfigDetails.UniqueClientIdentifier = data.UniqueClientIdentifier
				static.ConfigDetails.ThrottleValue[podName] = data.ThrottleValue
				static.ConfigDetails.Duration = data.Duration
				fmt.Println("Config set successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)

	}

	wg.Wait()

	return nil

}

func SetConfigInTraderPods(data static.Config, traderPodMap map[string]string, responses []map[string]interface{}) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for podName, podIP := range traderPodMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/setConfig"

		var traders []static.Trader

		if data.TargetEnv == 1 {
			traders = AssignMembersToPods(traders, static.LabSessionIdSets, podName, static.TRADER_MEMBERID)
		} else if data.TargetEnv == 2 {
			traders = AssignMembersToPods(traders, static.ProdSessionIdSets, podName, static.TRADER_MEMBERID)
		} else {
			traders = AssignMembersToPods(traders, static.SimSessionIdSets, podName, static.TRADER_MEMBERID)
		}

		// throttle := data.Trader_OPS / ((static.TRADER_CONNECTION_COUNT) / 2)
		throttle := data.Trader_OPS / static.TRADER_CONNECTION_COUNT
		fmt.Println("Trader throttle value: ", throttle)
		msg := static.Config{
			TotalOrderCount:        data.TotalOrderCount,
			TargetEnv:              data.TargetEnv,
			ModifyPercent:          data.ModifyPercent,
			CancelPercent:          data.CancelPercent,
			ThrottleValue:          throttle,
			FileType:               data.FileType,
			TraderCount:            static.TRADER_CONNECTION_COUNT,
			HeartbeatValue:         data.HeartbeatValue,
			UniqueClientIdentifier: data.UniqueClientIdentifier,
			Traders:                traders,
			TradeFlag:              1,
			Duration:               data.Duration,
			ProcessType:            data.ProcessType,
			AppSystemVendorName:    data.AppSystemVendorName,
			AppSystemVersion:       data.AppSystemVersion,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error : Error marshalling request - %v", err.Error())
		}

		wg.Add(1)
		go func(addr string, podName string) {
			defer wg.Done()
			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("POST", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in setting config for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in setting config for " + podName + " : " + err.Error())
						mu.Unlock()
						return
					} else {
						count -= 1
						continue
					}
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Println("Error in http.client")
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in setting config for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in setting config for " + podName + " : " + err.Error())
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
					fmt.Println("[ERR] Error in setting config in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in setting config for " + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})

				fmt.Println("Config Set Response from Trader Pods: ", responses)

				static.EventChannel <- static.SseData{Message: "Config set successfully for " + podName, Type: 0}

				static.ConfigDetails.TotalOrderCount = data.TotalOrderCount
				static.ConfigDetails.TargetEnv = data.TargetEnv
				static.ConfigDetails.ModifyPercent = data.ModifyPercent
				static.ConfigDetails.CancelPercent = data.CancelPercent
				static.ConfigDetails.FileType = data.FileType
				static.ConfigDetails.TraderCount = data.TraderCount
				static.ConfigDetails.HeartbeatValue = data.HeartbeatValue
				static.ConfigDetails.UniqueClientIdentifier = data.UniqueClientIdentifier
				static.ConfigDetails.ThrottleValue[podName] = data.ThrottleValue
				fmt.Println("Config set successfully for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()
	return nil
}

func DeleteOrdersInSenderPods(data static.DeleteRecordStruct, senderPodMap map[string]string, responses []map[string]interface{}) error {

	var wg sync.WaitGroup
	var mu sync.Mutex
	for podName, podIP := range senderPodMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/deleteRecords"

		msg := static.DeleteRecordStruct{
			InstId: data.InstId,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error : Error marshalling request - %v", err.Error())
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
					static.EventChannel <- static.SseData{Message: "Error in deleting record for" + podName, Type: 1}
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

	return nil

}

func SetOPSRateInPods(opsRateData static.OPSRate) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)
	for _, pod := range opsRateData.Pods {
		var podName string
		var podIP string
		if opsRateData.Type == 0 {
			podName = static.SENDER_POD_LABEL + "-" + strconv.Itoa(pod)
			podIP = static.SenderPodIPMap[podName]
		} else if opsRateData.Type == 1 {
			podName = static.TRADER_POD_LABEL + "-" + strconv.Itoa(pod)
			podIP = static.TraderPodIPMap[podName]
		}

		fmt.Println("In Change OPS, Pod Name: ", podName, " and Pod IP: ", podIP)
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/changeOPS"

		msg := static.OPSRate{
			Throttle: opsRateData.Throttle,
			Type:     opsRateData.Type,
		}

		reqMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error marshalling request %v", err.Error())
		}

		wg.Add(1)
		go func(addr string, podName string) {
			defer wg.Done()
			count := static.MaxRetry
			for count > 0 {
				resp, err := http.Post(addr, "application/json", bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in changing OPS rate for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in changing OPS rate for " + podName + " : " + err.Error())
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
					fmt.Println("[ERR] Error in changing OPS rate in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in changing OPS rate for" + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "OPS rate changed to " + strconv.Itoa(opsRateData.Throttle) + " for " + podName, Type: 0}
				static.ConfigDetails.ThrottleValue[podName] = opsRateData.Throttle
				fmt.Println("OPS rate changed to " + strconv.Itoa(opsRateData.Throttle) + " for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	return nil
}

func SendPatternGenerateDataToPods(generatePatternData static.GraphPattern, podLabel string) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)
	for _, number := range generatePatternData.Pods {
		var podIP string
		podName := podLabel + "-" + strconv.Itoa(number)
		if podLabel == static.SENDER_POD_LABEL {
			podIP = static.SenderPodIPMap[podName]
			// } else if podLabel == static.PEAK_POD_LABEL {
			// 	podIP = static.PeakPodIPMap[podName]
		} else {
			return fmt.Errorf("invalid pod label")
		}
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/generatePattern"
		msg := static.GraphPattern{
			Min:      generatePatternData.Min,
			Max:      generatePatternData.Max,
			Type:     generatePatternData.Type,
			Step:     generatePatternData.Step,
			Interval: generatePatternData.Interval,
			DelayMin: generatePatternData.DelayMin,
			DelayMax: generatePatternData.DelayMax,
			// ProcessType:         generatePatternData.ProcessType,
			// AppSystemVendorName: generatePatternData.AppSystemVendorName,
			// AppSystemVersion:    generatePatternData.AppSystemVersion,
		}
		reqMsg, err := json.Marshal(msg)

		if err != nil {
			return fmt.Errorf("error marshalling request: %v", err.Error())
		}

		wg.Add(1)
		go func(addr string, podName string) {
			defer wg.Done()
			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("POST", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in generating pattern for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in generating pattern for " + podName + " : " + err.Error())
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
						static.EventChannel <- static.SseData{Message: "Error in generating pattern for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in generating pattern for " + podName + " : " + err.Error())
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
					fmt.Println("[ERR] Error while setting generate pattern in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in generating pattern for " + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Successfully Generating pattern for " + podName, Type: 0}
				fmt.Println("Successfully Generating pattern for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	return nil
}

func GetActivePodIds(label string) []int {
	var activeSenderPods []int
	for _, value := range static.MemberPodMap {
		if strings.Contains(value, label) {
			podId, err := ExtractPodId(value)
			if err != nil {
				fmt.Println(err.Error())
			}
			activeSenderPods = append(activeSenderPods, podId)
		}
	}
	return activeSenderPods
}

func ExtractPodId(podName string) (int, error) {
	parts := strings.Split(podName, "-")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid pod name format")
	}

	lastPart := parts[len(parts)-1]
	number, err := strconv.Atoi(lastPart)
	if err != nil {
		return 0, fmt.Errorf("last part is not a number: %w", err)
	}

	return number, nil

}

func SetSlidingPriceInPods(slidingPriceData static.PriceRangeChangeDetails) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	responses := make([]map[string]interface{}, 0)

	for podName, podIP := range static.SenderPodIPMap {
		podAddr := "http://" + podIP + ":" + static.ServicePort + "/api/v1/addSlidingPriceRange"

		msg := static.PriceRangeChangeDetails{
			Interval:    slidingPriceData.Interval,
			Instruments: slidingPriceData.Instruments,
		}
		reqMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error marshalling request: %v", err.Error())
		}

		wg.Add(1) // Increment the wait group counter
		go func(addr string, podName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			count := static.MaxRetry

			for count > 0 {
				req, err := http.NewRequest("POST", addr, bytes.NewBuffer([]byte(reqMsg)))
				if err != nil {
					if count == 1 {
						mu.Lock()
						responses = append(responses, map[string]interface{}{
							"name":  podName,
							"error": err.Error(),
						})
						static.EventChannel <- static.SseData{Message: "Error in changing price details for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in changing price details for " + podName + " : " + err.Error())
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
						static.EventChannel <- static.SseData{Message: "Error in changing price details for " + podName + " : " + err.Error(), Type: 1}
						fmt.Println("Error in changing price details for " + podName + " : " + err.Error())
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
					fmt.Println("[ERR] Error in changing price details in pod: ", responseBody)
					static.EventChannel <- static.SseData{Message: "Error in changing price details for" + podName, Type: 1}
					continue
				}

				responses = append(responses, map[string]interface{}{
					"name":     podName,
					"response": string(body),
				})
				static.EventChannel <- static.SseData{Message: "Successfully changed price details for " + podName, Type: 0}
				fmt.Println("Successfully changed price details for " + podName)
				mu.Unlock()
				break
			}
		}(podAddr, podName)
	}

	wg.Wait()

	return nil
}

func AddPodsWrapper(reqPodCount int) (map[string]string, error) {

	podsStatusFlag, inactivePodsArr := GetPumpingStatusOfPods()

	availableMemberCount := CheckAvailableMemberIds()
	fmt.Println("Available Member IDs: ", availableMemberCount)
	if availableMemberCount == 0 {
		return nil, fmt.Errorf("no session ids left. cannot create pods")
	}
	possiblePodCount := availableMemberCount
	if reqPodCount > possiblePodCount {
		fmt.Printf("only %d pod(s) will be created\n", possiblePodCount)
		reqPodCount = possiblePodCount
	}
	podIPMap := make(map[string]string)
	if podsStatusFlag && len(inactivePodsArr) > 0 {
		existingPodStatus, totalPodsStarted, newPodMap := UseExistingPodsInStatefulSet(reqPodCount)
		maps.Copy(podIPMap, newPodMap)
		if existingPodStatus && totalPodsStarted == reqPodCount {
			fmt.Printf("%d pod(s) activated\n", reqPodCount)
			return newPodMap, nil
		} else if existingPodStatus && totalPodsStarted < reqPodCount {
			reqPodCount -= totalPodsStarted
		}
	}

	newPodMap, err := UseNewPodsInStatefulSet(reqPodCount)
	if err != nil {
		return nil, fmt.Errorf("error creating new %v pod: %v", reqPodCount, err.Error())
	}
	maps.Copy(podIPMap, newPodMap)

	return podIPMap, nil

}

func RestartPod(podName string) error {
	clientset, err := GetK8sClientset()
	if err != nil {
		return err
	}
	if err = DeletePodInStatefulSet(clientset, static.K8S_NAMESPACE, podName); err != nil {
		return fmt.Errorf("unable to delete pod %s: %v", podName, err.Error())
	}

	totalSenderPods, err := GetTotalPodCountInStatefulSet(clientset, static.K8S_NAMESPACE, static.SENDER_POD_LABEL)
	if err != nil {
		return fmt.Errorf("unable to get total pod count: %v", err.Error())
	}

	if err := CreatePodMap(int(totalSenderPods)); err != nil {
		return fmt.Errorf("unable to unable to create pod map: %v", err.Error())
	}
	return nil
}
