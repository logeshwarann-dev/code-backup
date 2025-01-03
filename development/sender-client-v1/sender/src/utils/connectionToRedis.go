package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/go-redis/redis/v8"
)

func RedisConnection() (*redis.Client, error) {

	secondaryRedisURL := getSecondaryRedisPodAddr()

	rdb := redis.NewClient(&redis.Options{
		// Addr: "localhost:6379",
		Addr: secondaryRedisURL,
		DB:   DB_FOR_CONFIG,
	})

	// Test connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis (DB: %d): %v", DB_FOR_CONFIG, err)
	}

	fmt.Printf("Connected to Redis (DB: %d)\n", DB_FOR_CONFIG)
	return rdb, nil
}

func getSecondaryRedisPodAddr() string {
	podName := os.Getenv("POD_NAME")
	podIndex, _ := strconv.Atoi(strings.Split(podName, "-")[1]) // Extract pod index from pod name (e.g., client-0 -> 0)

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
		LabelSelector: "app=secondary-redis",
	})
	if err != nil {
		log.Fatalf("Error getting secondary Redis pods: %v", err)
	}

	podIP := pods.Items[podIndex].Status.PodIP
	return fmt.Sprintf("%s:6379", podIP)
}
