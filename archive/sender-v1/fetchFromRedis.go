package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"math/rand"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func init() {

	rdb, err := RedisConnection()
	if err != nil {
		fmt.Println("Error While connecting to redis")
		return
	}

	err = getDataFromLocalRedis(rdb)
	if err != nil {
		fmt.Println("Error in detching data from local redis", err)
	}

}

func getDataFromLocalRedis(rdb *redis.Client) error {

	parsedTimeStart, err := time.Parse(LAYOUT, START_TIME_STR)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		return err
	}

	utcTime := parsedTimeStart.UTC()
	startTime := utcTime.Unix()

	parsedTimeEnd, err := time.Parse(LAYOUT, END_TIME_STR)
	if err != nil {
		fmt.Println("Error parsing end time:", err)
		return err
	}

	utcTime = parsedTimeEnd.UTC()
	endtime := utcTime.Unix()

	diff := int(endtime - startTime)
	TOTAL_RUN_TIME = diff

	if diff%10 == 0 {
		diff = diff / 10
		diff--
	} else {

		diff = diff / 10

	}

	for i := 0; i <= diff; i++ {

		fmt.Println("Iteration of diff:", i, diff)
		jsonData, err := rdb.Get(ctx, strconv.Itoa(i)).Result()
		if err != nil {
			return fmt.Errorf("error retrieving data: %v", err)
		}

		var escapedData []interface{} // Change to interface{} to support both strings and arrays
		err = json.Unmarshal([]byte(jsonData), &escapedData)
		if err != nil {
			fmt.Printf("Error unmarshaling raw data: %v\n", err)
			return err
		}

		// Step 2: Iterate over the slice and unmarshal each item after removing backslashes
		var parsedData [][]string
		for _, item := range escapedData {
			// Handle both string and array types
			switch v := item.(type) {
			case string:
				// Remove escape characters (backslashes) from string
				cleanedItem := strings.Replace(v, "\\", "", -1)

				var innerData []string
				err := json.Unmarshal([]byte(cleanedItem), &innerData)
				if err != nil {
					fmt.Printf("Error unmarshaling inner data: %v\n", err)
					return err
				}
				parsedData = append(parsedData, innerData)
			case []interface{}:
				// Handle the case where item is an array
				var innerData []string
				for _, innerItem := range v {
					if innerStr, ok := innerItem.(string); ok {
						innerData = append(innerData, innerStr)
					}
				}
				parsedData = append(parsedData, innerData)
			default:
				fmt.Printf("Unexpected item type: %T\n", v)
			}
		}

		for _, item := range parsedData {
			// Send each item, including empty arrays, to the main channel
			mainChannel <- item
		}
		// Print the cleaned data
		// fmt.Println("Parsed Data:", parsedData)
	}

	return nil
}

func ProcessValues(value []string) []redisData {
	var dataItems []redisData

	// Iterate over each element in the input slice
	for _, groupStr := range value {
		var group []string

		// Parse the string as JSON array of strings
		err := json.Unmarshal([]byte(groupStr), &group)
		if err != nil {
			fmt.Printf("Error unmarshalling group: %v", err)
			continue
		}

		// Process each entry in the parsed group
		for _, entry := range group {
			fields := strings.Split(entry, ",")
			if len(fields) != 11 {
				fmt.Printf("Invalid entry format: %s", entry)
				continue
			}

			// Parse fields
			templateID, _ := strconv.Atoi(fields[0])
			instID, _ := strconv.Atoi(fields[1])
			price, _ := strconv.ParseFloat(fields[2], 64)
			orderQty, _ := strconv.Atoi(fields[3])
			maxShow, _ := strconv.Atoi(fields[4])
			accType := fields[5]
			timeInForce := fields[6]
			clientCode := fields[7]
			msgSeqNum, _ := strconv.Atoi(fields[8])
			transactionType, _ := strconv.Atoi(fields[9])
			orderID, _ := strconv.ParseFloat(fields[10], 64)

			randomSession := rand.Intn(103401050-103401001+1) + 103401001
			randomSession = 103401001

			// Create the redisData item
			dataItem := redisData{
				TemplateID:      templateID,
				Price:           price,
				OrderQty:        orderQty,
				MaxShow:         maxShow,
				InstID:          instID,
				AccType:         accType,
				TimeInForce:     timeInForce,
				ClientCode:      clientCode,
				MsgSeqNum:       msgSeqNum,
				TraderId:        randomSession,
				PartitionId:     0,
				ProductId:       0,
				timeStamp:       0,
				TransactionType: transactionType,
				orderID:         orderID,
			}

			dataItems = append(dataItems, dataItem)
		}
	}

	return dataItems
}
