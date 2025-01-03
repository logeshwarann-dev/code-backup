package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var ctx = context.Background()

func SetConfig(db *redis.Client) error {
	hashKey := "config"

	// Marshal the struct into a map
	data, err := json.Marshal(CONFIG_VALUES)
	if err != nil {
		fmt.Printf("Error marshaling CONFIG_VALUES: %v\n", err)
		return err
	}

	var configMap map[string]interface{}
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		fmt.Printf("Error unmarshaling CONFIG_VALUES: %v\n", err)
		return err
	}

	// Convert values to strings and prepare for HSet
	redisData := make(map[string]string, len(configMap))
	for key, value := range configMap {
		redisData[key] = fmt.Sprintf("%v", value)
	}

	// Use HSet to store all fields in one call
	err = db.HSet(ctx, hashKey, redisData).Err()
	if err != nil {
		fmt.Printf("Error setting config in Redis: %v\n", err)
		return err
	}

	fmt.Println("Config stored in Redis hash successfully.")
	return nil
}

// ============MASTER REDIS=========

func SetConstants(db *redis.Client) error {

	// config, err := db.HGetAll(ctx, "config").Result()
	// if err != nil {
	// 	return fmt.Errorf("error fetching config: %w", err)
	// }

	START_TIME_STR = CONFIG_VALUES.StartTime
	END_TIME_STR = CONFIG_VALUES.EndTime

	if CONFIG_VALUES.MarketReplay == 1 {
		MARKET_REPLAY = true
	} else {
		MARKET_REPLAY = false
	}

	fmt.Printf("Start Time: %s, End Time: %s\n", START_TIME_STR, END_TIME_STR)

	memberSet, err := db.HGetAll(ctx, "member_set").Result()
	if err != nil {
		return fmt.Errorf("error fetching memberSet: %w", err)
	}

	groupSize := CONFIG_VALUES.Traders
	for i, key := range sortedKeys(memberSet) {
		groupName := fmt.Sprintf("set%d", i/groupSize)
		if MEMBER_SET[groupName] == nil {
			MEMBER_SET[groupName] = make(map[string]string)
		}
		MEMBER_SET[groupName][key] = memberSet[key]
	}

	fmt.Println("Length of MemberSEt: ", len(MEMBER_SET))

	if !MARKET_REPLAY {

		instruments, err := db.HGetAll(ctx, "ORDER_PUMPING").Result()
		if err != nil {
			return fmt.Errorf("error fetching memberSet: %w", err)
		}

		if len(CONFIG_VALUES.Instruments) != 0 {
			for _, key := range CONFIG_VALUES.Instruments {
				strKey := strconv.Itoa(key)
				if value, exists := instruments[strKey]; exists {
					INSTRUMENTS[strKey] = value
				}
			}
		} else {
			INSTRUMENTS = instruments
		}

		fmt.Println("INSTRUMENTS :", INSTRUMENTS)
	}

	return nil
}

// ==========SECONDARY REDIS PODS===========

func (db *DB) SetDataToSecondaryRedis(redisPodId int) error {

	if err := SetConfig(db.traders_db); err != nil {
		fmt.Println("Error in HSET of secondary db:", err)
		return err
	}

	if !MARKET_REPLAY {
		if err := setInstrumentsInSecondaryRedis(db.traders_db, redisPodId); err != nil {
			fmt.Println("Error in HSET of secondary db:", err)
			return err
		}
	}
	if err := SetMembersetInSecondaryRedis(db.traders_db, redisPodId); err != nil {
		fmt.Println("Error in HSET of secondary db:", err)
		return err
	}

	if err := SwitchDatabase(db.main_db, 5); err != nil {
		fmt.Println("Failed to switch database:", err)
	}

	if err := SwitchDatabase(db.traders_db, 1); err != nil {
		fmt.Println("Failed to switch database:", err)
	}

	parsedTimeStart, err := time.Parse(LAYOUT, START_TIME_STR)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		return err
	}

	utcTime := parsedTimeStart.UTC()
	initial_second := utcTime.Unix()
	final_second := initial_second + BATCH_SIZE

	parsedTimeEnd, err := time.Parse(LAYOUT, END_TIME_STR)
	if err != nil {
		fmt.Println("Error parsing end time:", err)
		return err
	}

	utcTime = parsedTimeEnd.UTC()
	endtime := utcTime.Unix()

	for {

		if initial_second == endtime {
			err = db.fetchAndParseRedisData(initial_second, endtime)
			if err != nil {
				fmt.Println("Error in case where start time is equal to end time:", err)
				return err
			}
			return nil
		} else if endtime-initial_second < BATCH_SIZE {
			err = db.fetchAndParseRedisData(initial_second, endtime)
			if err != nil {
				fmt.Printf("Error in case where diff between start time & end time is less than %d: %v \n", BATCH_SIZE, err)
				return err
			}
			return nil
		} else if initial_second < endtime {

			err = db.fetchAndParseRedisData(initial_second, final_second)
			if err != nil {
				fmt.Println("Error in case where current time is less than end time:", err)
			}

			initial_second = final_second + 1
			final_second = initial_second + BATCH_SIZE

			time.Sleep(1000 * time.Millisecond)
		}

	}

}

func SetMembersetInSecondaryRedis(db *redis.Client, podId int) error {

	hashKey := "member_set"
	groupName := fmt.Sprintf("set%d", podId)
	memberset_data := MEMBER_SET[groupName]

	err := db.HSet(ctx, hashKey, memberset_data).Err()
	if err != nil {
		fmt.Printf("Error setting config in Secondary Redis: %v\n", err)
		return err
	}
	fmt.Printf("Memberset[%v] stored in Secondary-Redis-%d hash successfully.\n", groupName, podId)
	return nil

}

func setInstrumentsInSecondaryRedis(db *redis.Client, podId int) error {
	hashKey := "instruments"
	err := db.HSet(ctx, hashKey, INSTRUMENTS).Err()
	if err != nil {
		fmt.Printf("Error setting config in Secondary Redis: %v\n", err)
		return err
	}
	fmt.Printf("Instruments data stored in Secondary-Redis-%d hash successfully.\n", podId)
	return nil
}

// ============Data Operations========

func (db *DB) fetchAndParseRedisData(startTime, endTime int64) error {

	var allKeys [][]string

	starttime := time.Now()

	fmt.Printf("Set: %d Start Time : %d End Time: %d \n", db.Set, startTime, endTime)
	id := "set" + strconv.Itoa(db.Set)
	db.session_ids = MEMBER_SET[id]

	for currentTime := startTime; currentTime <= endTime; currentTime++ {
		var timeKeys []string
		for _, sessionID := range db.session_ids {

			var result map[string]interface{}
			if err := json.Unmarshal([]byte(sessionID), &result); err != nil {
				log.Fatalf("Error unmarshaling JSON: %v", err)
			}

			if id, ok := result["session_id"].(float64); ok {
				s_ID := int(id)
				pattern := fmt.Sprintf("%d:%d:*:*", currentTime, s_ID)
				timeKeys = append(timeKeys, pattern)
			} else {
				log.Println("session_id not found or invalid")
			}

		}
		allKeys = append(allKeys, timeKeys)
	}

	dataItems, err := db.fetchDataByPatterns(ctx, allKeys)
	if err != nil {
		log.Printf("Error fetching data: %v\n", err)
		return err
	}

	if len(dataItems) == 0 {
		fmt.Println("No data to store.")
		return nil
	}

	// Serialize the 2D array into JSON
	dataJSON, err := json.Marshal(dataItems)
	if err != nil {
		log.Printf("Error serializing data items: %v\n", err)
		return err
	}

	// Store the serialized data under a single key in db1
	err = db.traders_db.Set(ctx, strconv.Itoa(db.key), dataJSON, 0).Err()
	if err != nil {
		log.Printf("Error setting data in traders db of Set '%d': %v\n", db.Set, err)
		return err
	}

	fmt.Printf("Successfully stored data under key '%d' in db.\n", db.key)
	db.key++

	fmt.Println("Time took to write data to sub redis:", time.Since(starttime))

	return nil
}

func (db *DB) fetchDataByPatterns(ctx context.Context, patterns [][]string) ([]interface{}, error) {
	var allResults []interface{}

	for _, patternGroup := range patterns {
		var allMatchingKeys []string

		for _, pattern := range patternGroup {
			keys, err := db.main_db.Keys(ctx, pattern).Result()
			if err != nil {
				return nil, fmt.Errorf("error fetching keys for pattern '%s': %v", pattern, err)
			}

			allMatchingKeys = append(allMatchingKeys, keys...)
		}

		if len(allMatchingKeys) == 0 {
			// Append an empty slice when no keys are found
			allResults = append(allResults, []interface{}{})
		} else {
			values, err := db.main_db.MGet(ctx, allMatchingKeys...).Result()
			if err != nil {
				return nil, fmt.Errorf("error fetching values for keys: %v", err)
			}

			_, err = db.main_db.Del(ctx, allMatchingKeys...).Result()
			if err != nil {
				return nil, fmt.Errorf("error deleting keys: %v", err)
			}

			allResults = append(allResults, values)
		}
	}

	return allResults, nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

func SwitchDatabase(db *redis.Client, dbNumber int) error {
	if err := db.Do(ctx, "SELECT", dbNumber).Err(); err != nil {
		return fmt.Errorf("error switching to DB %d: %w", dbNumber, err)
	}
	return nil
}

func GetRedisPodList(desiredCount int) []v1.Pod {

	var podsList []v1.Pod
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating Kubernetes config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	for try := range 5 {
		// Get list of secondary Redis pods
		time.Sleep(5 * time.Second)
		pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app=secondary-redis",
		})
		if err != nil {
			log.Fatalf("Error getting secondary Redis pods: %v", err)
		}
		podsList = pods.Items

		if len(podsList) == desiredCount {
			break
		}
		fmt.Printf("Pods Count: %d || Retrying to fetch pods - attempt %v\n", len(podsList), try)
	}

	return podsList
}

func ScaleRedisAndSenderPods(targetCount int /*context *gin.Context*/) error {
	// var request struct {
	// 	ReplicaCount int `json:"replica_count"`
	// }
	// if err := context.ShouldBindJSON(&request); err != nil {
	// 	context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	// 	return
	// }
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Scale secondary redis and sender clients
	namespace := "default"
	statefulSets := []string{"secondary-redis", "client"}

	for _, statefulSet := range statefulSets {
		scaleErr := ScaleStatefulSet(clientset, namespace, statefulSet, targetCount)
		if scaleErr != nil {
			return fmt.Errorf("pod Scaling Error: %v", scaleErr.Error())
		}
	}

	return nil
	// context.JSON(http.StatusOK, gin.H{"message": "Redis and Trader Client pods are scaled successfully"})

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

	log.Printf("Scaled StatefulSet %s to %d replicas\n", statefulSetName, replicas)
	return nil
}

func constructTradersDB(redisPods []v1.Pod) {

	if len(Traders_DB) > 1 {
		clearTradersDB(&Traders_DB)
	}
	for i := range len(MEMBER_SET) {
		podIP := redisPods[i].Status.PodIP
		secondaryRedisURL := fmt.Sprintf("%s:6379", podIP)
		setName := fmt.Sprintf("set%d", i)
		Traders_DB = append(Traders_DB, DB{Set: i, Addr: secondaryRedisURL, session_ids: MEMBER_SET[setName]})
	}
	// {set: 0, Addr: "old addr",session_id: "" }... {set: 7, Adrr: "old addr"}, {set: 8, Addr: "new addr"}
}

func clearTradersDB(db *[]DB) {
	// Reassign the slice to an empty slice
	*db = []DB{}
}

// ===========ARCHIVE=================

// func FetchMembersFromMainDb(db *redis.Client) error {

// 	// config, err := db.HGetAll(ctx, "config").Result()
// 	// if err != nil {
// 	// 	return fmt.Errorf("error fetching config: %w", err)
// 	// }

// 	START_TIME_STR = CONFIG_VALUES.StartTime
// 	END_TIME_STR = CONFIG_VALUES.EndTime

// 	fmt.Printf("Start Time: %s, End Time: %s\n", START_TIME_STR, END_TIME_STR)

// 	memberSet, err := db.HGetAll(ctx, "memberSet").Result()
// 	if err != nil {
// 		return fmt.Errorf("error fetching memberSet: %w", err)
// 	}

// 	groupSize := CONFIG_VALUES.Traders
// 	for i, key := range sortedKeys(memberSet) {
// 		groupName := fmt.Sprintf("set%d", i/groupSize)
// 		if MEMBER_SET[groupName] == nil {
// 			MEMBER_SET[groupName] = make(map[string]string)
// 		}
// 		MEMBER_SET[groupName][key] = memberSet[key]
// 	}
// 	return nil
// }
