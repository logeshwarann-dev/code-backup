package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type TradersList struct {
	MemberIds []string `json:"member_ids"`
}

var memberSet = map[string]int{
	"1034": 0,
	"1035": 1,
	"1036": 2,
	"1037": 3,
}

func main() {

	server := gin.Default()

	server.POST("/api/start-pumping", triggerSender)

	server.Run(":8085")

}

func triggerSender(ctx *gin.Context) {
	var traderList TradersList
	podsList := getSecondaryRedisPodAddr()
	if err := ctx.BindJSON(&traderList); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, memberId := range traderList.MemberIds {
		clientPod := memberSet[memberId]
		senderPodIP := podsList.Items[clientPod].Status.PodIP
		senderAddr := fmt.Sprintf("http://%s:8090/go-sender", senderPodIP)
		resp, err := http.Post(senderAddr, "application/json", bytes.NewBuffer([]byte(`{"message": "Start Pumping!!"}`)))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()
		response, err := io.ReadAll(resp.Body)
		ctx.JSON(http.StatusOK, gin.H{"status": response})
	}

}

func getSecondaryRedisPodAddr() *v1.PodList {

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
		LabelSelector: "app=client",
	})
	if err != nil {
		log.Fatalf("Error getting client pods: %v", err)
	}

	return pods

}
