package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	// "github.com/segmentio/kafka-go"
)

var ctx = context.Background()

func (db *DB) SetDataToSecondaryRedis() error {

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

func (db *DB) fetchAndParseRedisData(startTime, endTime int64) error {

	var allKeys [][]string

	fmt.Printf("Set: %d Start Time : %d End Time: %d \n", db.Set, startTime, endTime)

	for currentTime := startTime; currentTime <= endTime; currentTime++ {
		var timeKeys []string
		for _, sessionID := range db.session_ids {
			pattern := fmt.Sprintf("%d:%d:*:*", currentTime, sessionID)
			timeKeys = append(timeKeys, pattern)
		}
		allKeys = append(allKeys, timeKeys)
	}

	// fmt.Println(allKeys, "ALL KEYS>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	dataItems, err := db.fetchDataByPatterns(ctx, allKeys)
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		return err
	}

	fmt.Printf("Fetched records between timestamps %d & %d : %v \n", startTime, endTime, dataItems)

	if len(dataItems) == 0 {
		fmt.Println("No data to store.")
		return nil
	}

	// Serialize the 2D array into JSON
	dataJSON, err := json.Marshal(dataItems)
	if err != nil {
		log.Printf("Error serializing data items: %v", err)
		return err
	}

	// Store the serialized data under a single key in db1
	err = db.traders_db.Set(ctx, strconv.Itoa(db.key), dataJSON, 0).Err()
	if err != nil {
		log.Printf("Error setting data in traders db of Set '%d': %v", db.Set, err)
		return err
	}

	fmt.Printf("Successfully stored data under key '%d' in db.\n", db.key)
	db.key++

	// for i := 0; i < 100; i++ {

	// 	writeMessage(dataJSON)
	// }

	return nil
}

// func writeMessage(message []byte) {
// 	// Create a Kafka writer
// 	writer := kafka.Writer{
// 		Addr:     kafka.TCP(KAFKA_BROKER),
// 		Topic:    KAFKA_TOPIC,
// 		Balancer: &kafka.LeastBytes{},
// 	}

// 	// Use a context for the write operation
// 	ctx := context.Background()

// 	// Write the message to the topic
// 	err := writer.WriteMessages(ctx, kafka.Message{
// 		Key:   []byte("Key"),
// 		Value: message,
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to write message: %v", err)
// 	}

// 	fmt.Printf("Message written to %s: %s\n", KAFKA_TOPIC, message)
// }

func (db *DB) fetchDataByPatterns(ctx context.Context, patterns [][]string) ([]interface{}, error) {
	var allResults []interface{}

	for _, patternGroup := range patterns {
		var allMatchingKeys []string

		// Iterate over each pattern in the subarray
		for _, pattern := range patternGroup {
			cursor := uint64(0)
			// fmt.Printf("Scanning for pattern: %s\n", pattern)

			for {
				// Perform the scan for the current pattern
				scannedKeys, nextCursor, err := db.main_db.Scan(ctx, cursor, pattern, 1000).Result()
				if err != nil {
					return nil, fmt.Errorf("error scanning keys of Set '%d' for pattern %s: %v", db.Set, pattern, err)
				}

				// Append the scanned keys to the list
				allMatchingKeys = append(allMatchingKeys, scannedKeys...)

				// Update cursor and check if scanning is complete
				cursor = nextCursor
				if cursor == 0 {
					break
				}
			}
		}

		// Fetch values for all matching keys in this pattern group
		if len(allMatchingKeys) == 0 {
			// Append an empty slice when no keys are found
			allResults = append(allResults, []interface{}{})
		} else {
			values, err := db.main_db.MGet(ctx, allMatchingKeys...).Result()
			if err != nil {
				return nil, fmt.Errorf("error fetching values for Set '%d': %v", db.Set, err)
			}

			allResults = append(allResults, values)
		}
	}

	return allResults, nil
}
