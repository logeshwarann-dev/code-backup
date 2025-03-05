package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type UserRequest struct {
	PodCount int `json:"pod_count"`
}

type PodResponse struct {
	Address  string `json:"address"`
	Response string `json:"response"`
}

var (
	NODE1      = "bseitrspoc4.bseindia.com"
	NODE2      = "bseitrspoc5.bseindia.com"
	NODE3      = "bseitrspoc3.bseindia.com"
	NODE4      = "bseitrspoc2.bseindia.com"
	PODS_COUNT = 1
)

func DistributePods(c *gin.Context) {

	namespace := "default"
	statefulSetName := "mock-sender"
	// replicas := 3
	var userRequest UserRequest
	nodesArr := []string{NODE1, NODE2, NODE3, NODE4}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "request not in proper format" + err.Error(),
		})
		return
	}

	clientset, err := CreateK8sClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to create k8s client" + err.Error(),
		})
		return
	}

	nodesList, err := GetAllNodesInTheCluster(clientset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to get all nodes in cluster" + err.Error(),
		})
		return
	}
	allNodesActive := VerifyActiveNodes(nodesArr, nodesList)
	if !allNodesActive {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "all nodes are not active in cluster",
		})
		return
	}

	err = UpdateNodePodMap(clientset, namespace, statefulSetName, NodeAndPodMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to update node map in cluster" + err.Error(),
		})
		return
	}

	for range userRequest.PodCount {
		for _, desiredNode := range nodesArr {
			if len(NodeAndPodMap[desiredNode]) <= len(NodeAndPodMap[NODE4]) {
				PODS_COUNT++
				desiredPodCount := PODS_COUNT
				err = CreatePodInSpecificNode(clientset, namespace, statefulSetName, desiredPodCount, desiredNode)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Error while generating sender pods" + err.Error(),
					})
					fmt.Println("Error while generating sender pods" + err.Error())
					return
				}
				err = UpdateNodePodMap(clientset, namespace, statefulSetName, NodeAndPodMap)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "unable to update node map in cluster" + err.Error(),
					})
					return
				}
				break
			}

		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pods are scaled"})
}

func CreatePodInSpecificNode(clientset *kubernetes.Clientset, namespace, statefulSetName string, replicas int, nodeName string) error {

	statefulSet, err := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	newReplicaCount := int32(replicas)
	statefulSet.Spec.Replicas = &newReplicaCount

	statefulSet.Spec.Template.Spec.Affinity = &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{nodeName},
							},
						},
					},
				},
			},
		},
	}

	// Update StatefulSet
	_, err = clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Scaled StatefulSet %s to %d replicas with even distribution\n", statefulSetName, replicas)
	return nil
}

func UpdateNodePodMap(clientset *kubernetes.Clientset, namespace, statefulSetName string, nodePodMap map[string][]string) error {
	// time.Sleep(5 * time.Second)
	GetPodHealthStatus(statefulSetName, PODS_COUNT)
	pods, err := GetAllPodsInStatefulSet(clientset, namespace, statefulSetName)
	if err != nil {
		return err
	}
	clear(nodePodMap)
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if nodeName != "" {
			nodePodMap[nodeName] = append(nodePodMap[nodeName], pod.Name)
		}
	}
	fmt.Println("Updated Node-Pod Map:", nodePodMap)
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

func CreateK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
		return nil, err
	}

	return clientset, nil
}

func GetAllPodsInStatefulSet(clientset *kubernetes.Clientset, namespace string, statefulSetName string) ([]v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", statefulSetName),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func VerifyActiveNodes(nodeArr []string, nodeList []string) bool {
	var nodeActive bool
	for _, node := range nodeArr {
		nodeActive = slices.Contains(nodeList, node)
		if !nodeActive {
			return nodeActive
		}
	}
	return true

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
		var podHealthResponse PodResponse
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
	podsList, err := GetPodAddr(appLabel, "default")
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

	clientset, err := CreateK8sClient()
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
