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

func ConnectToSecondaryRedisPod() error {

	rdb, err := RedisConnection()
	if err != nil {
		fmt.Println("Error While connecting to redis")
		return err
	}

	err = ReadConfigFromRedis(rdb)
	if err != nil {
		fmt.Println("Error in reading config data from local redis", err)
		return err
	}

	err = getDataFromLocalRedis(rdb)
	if err != nil {
		fmt.Println("Error in fetching data from local redis", err)
		return err
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
			if len(fields) != 16 {
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

			// Create the redisData item
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

func getDataFromLocalRedis(rdb *redis.Client) error {

	if err := rdb.Do(ctx, "SELECT", 1).Err(); err != nil {
		return fmt.Errorf("error switching to DB 1: %w", err)
	}

	// Parse and convert start and end times to Unix timestamps
	startTime, err := parseUnixTimestamp(START_TIME_STR)
	if err != nil {
		return fmt.Errorf("error parsing start time: %w", err)
	}

	endTime, err := parseUnixTimestamp(END_TIME_STR)
	if err != nil {
		return fmt.Errorf("error parsing end time: %w", err)
	}

	// Calculate total run time and number of iterations
	diff := int(endTime - startTime)
	TOTAL_RUN_TIME = diff
	iterations := diff / 10
	if diff%10 == 0 {
		iterations--
	}

	// Iterate and process data
	for i := 0; i <= iterations; i++ {
		fmt.Printf("Processing iteration %d of %d\n", i, iterations)

		jsonData, err := rdb.Get(ctx, strconv.Itoa(i)).Result()
		if err != nil {
			return fmt.Errorf("error retrieving data: %w", err)
		}

		parsedData, err := parseRedisData(jsonData)
		if err != nil {
			return fmt.Errorf("error parsing Redis data: %w", err)
		}

		// Send parsed data to the main channel
		for _, item := range parsedData {
			MainChannel <- item
		}

		// //Delete the key after fetching
		// err = rdb.Del(ctx, strconv.Itoa(i)).Err()
		// if err != nil {
		// 	fmt.Println("error deleting key:", err)
		// }

		// fmt.Printf("Deleted key %d from redis db \n", i)

		time.Sleep(time.Duration(REDIS_TIMESTAMP_BATCH/2) * time.Second)
	}

	return nil
}

func parseUnixTimestamp(timeStr string) (int64, error) {
	parsedTime, err := time.Parse(LAYOUT, timeStr)
	if err != nil {
		return 0, err
	}
	return parsedTime.UTC().Unix(), nil
}

func parseRedisData(jsonData string) ([][]string, error) {
	var escapedData []interface{}
	if err := json.Unmarshal([]byte(jsonData), &escapedData); err != nil {
		return nil, fmt.Errorf("error unmarshaling data: %w", err)
	}

	var parsedData [][]string
	for _, item := range escapedData {
		switch v := item.(type) {
		case string:
			cleanedItem := strings.ReplaceAll(v, "\\", "")
			var innerData []string
			if err := json.Unmarshal([]byte(cleanedItem), &innerData); err != nil {
				return nil, fmt.Errorf("error unmarshaling inner data: %w", err)
			}
			parsedData = append(parsedData, innerData)
		case []interface{}:
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

	return parsedData, nil
}
