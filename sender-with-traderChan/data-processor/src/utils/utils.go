package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/go-redis/redis/v8"
)

func RedisConnection() error {
	ctx := context.Background()

	//------------------------------------
	masterRedisIP := os.Getenv("master-redis")
	masterRedisURL := fmt.Sprintf("%v:6379", masterRedisIP)

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

	//-----------------------------------------------
	mainRdb := redis.NewClient(&redis.Options{
		//Addr: "localhost:6379",
		Addr: masterRedisURL,
		DB:   0,
	})

	if _, err := mainRdb.Ping(ctx).Result(); err != nil {
		fmt.Printf("Failed to connect to main Redis: %v\n", err)
		return err
	}

	err = FetchConfigFromMainDb(mainRdb)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup

	for i := range Traders_DB {
		wg.Add(1)

		database := &Traders_DB[i]
		go func(db *DB) {
			defer wg.Done()
			//----------------------------

			podIP := pods.Items[i].Status.PodIP
			secondaryRedisURL := fmt.Sprintf("%s:6379", podIP)

			//------------------------
			traderClient := redis.NewClient(&redis.Options{
				// Addr: db.Addr,
				//Addr: "localhost:6379",
				Addr: secondaryRedisURL,
				DB:   0,
			})

			if _, err := traderClient.Ping(ctx).Result(); err != nil {
				fmt.Printf("Failed to connect to trader Redis: %v\n", err)
				return
			}

			db.main_db = mainRdb
			db.traders_db = traderClient

			fmt.Println("REDIS CONNECTED SUCCESSFULLY FOR TRADER SET:", db.Set)

			keys := []string{"config", "member_set"}
			err := db.SetConfig(keys)
			if err != nil {
				fmt.Println("Error in assigning config to sub redis:", err)
			}

			err = db.FetchFromRedis()
			if err != nil {
				fmt.Println("Error in Fetch from Redis", err)
			}

		}(database)
	}

	wg.Wait()
	return nil
}
