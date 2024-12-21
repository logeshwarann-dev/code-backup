package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var MainChannel = make(chan []string, MAIN_CHANNEL_LENGTH)

func init() {

	rdb, err := RedisConnection()
	if err != nil {
		fmt.Println("Error While connecting to redis, ", err)
		return
	}

	err = ReadConfigFromRedis(rdb)
	if err != nil {
		fmt.Println("Error in reading config data from local redis", err)
		return
	}

	err = getDataFromLocalRedis(rdb)
	if err != nil {
		fmt.Println("Error in fetching data from local redis", err)
		return
	}

}

func getDataFromLocalRedis(rdb *redis.Client) error {

	if err := rdb.Do(ctx, "SELECT", 1).Err(); err != nil {
		fmt.Println("Error switching to DB 1:", err)
		return err
	}

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
			MainChannel <- item
		}
		// Print the cleaned data
		// fmt.Println("Parsed Data:", parsedData)
	}

	return nil
}

func ProcessValues(value []string) []RedisData {
	var dataItems []RedisData

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
			if len(fields) < 11 {
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
			timeStamp, _ := strconv.Atoi(fields[11])
			productID, _ := strconv.Atoi(fields[12])
			mID, _ := strconv.Atoi(fields[13])
			tID, _ := strconv.Atoi(fields[14])
			session_id := (mID * 100000) + tID
			patitionID, _ := strconv.Atoi(fields[15])

			// Create the RedisData item
			dataItem := RedisData{
				TemplateID:      templateID,
				Price:           price,
				OrderQty:        orderQty,
				MaxShow:         maxShow,
				InstID:          instID,
				AccType:         accType,
				TimeInForce:     timeInForce,
				ClientCode:      clientCode,
				MsgSeqNum:       msgSeqNum,
				TraderId:        session_id,
				PartitionId:     patitionID,
				ProductId:       productID,
				TimeStamp:       timeStamp,
				TransactionType: transactionType,
				OrderID:         orderID,
			}

			dataItems = append(dataItems, dataItem)
		}
	}

	return dataItems
}

func ReadConfigFromRedis(rdb *redis.Client) error {

	config, err := rdb.HGetAll(ctx, "config").Result()
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}

	memberSet, err := rdb.HGetAll(ctx, "member_set").Result()
	if err != nil {
		return fmt.Errorf("failed to fetch member_set: %v", err)
	}

	err = SetConfigValues(config, memberSet)
	if err != nil {
		fmt.Println("Error in setting config values:", err)
		return err
	}

	return nil
}
