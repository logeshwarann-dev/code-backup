package main

//Initial Implementation

// func (db *DB) fetchAndParseRedisData(startTime, endTime int64) error {

// 	var allKeys []string

// 	fmt.Printf("Set: %d Start Time : %d End Time: %d \n", db.Set, startTime, endTime)

// 	for currentTime := startTime; currentTime <= endTime; currentTime++ {
// 		for _, sessionID := range db.session_ids {
// 			pattern := fmt.Sprintf("%d:%d:*:*", currentTime, sessionID)
// 			allKeys = append(allKeys, pattern)
// 		}
// 	}

// 	fmt.Println(allKeys, "KEYSSS................")

// 	dataItems, err := db.fetchDataByPatterns(ctx, allKeys)
// 	if err != nil {
// 		log.Printf("Error fetching data: %v", err)
// 		return err
// 	}

// 	fmt.Println("dataItems", dataItems)

// 	if len(dataItems) == 0 {
// 		fmt.Println("No data to store.")
// 		return nil
// 	}

// 	// Serialize the 2D array into JSON
// 	dataJSON, err := json.Marshal(dataItems)
// 	if err != nil {
// 		log.Printf("Error serializing data items: %v", err)
// 		return err
// 	}

// 	// Store the serialized data under a single key in db1
// 	err = db.traders_db.Set(ctx, strconv.Itoa(db.key), dataJSON, 0).Err()
// 	if err != nil {
// 		log.Printf("Error setting data in db1: %v", err)
// 		return err
// 	}

// 	db.key++

// 	fmt.Printf("Successfully stored data under key '%d' in db1.\n", db.key)

// 	return nil
// }

// func (db *DB) fetchDataByPatterns(ctx context.Context, patterns []string) ([]interface{}, error) {

// 	var allMatchingKeys []string
// 	keyToPatternIndex := make(map[string]int)

// 	for patternIndex, pattern := range patterns {
// 		cursor := uint64(0)
// 		fmt.Printf("Scanning for pattern: %s\n", pattern) // Debug log

// 		for {

// 			scannedKeys, nextCursor, err := db.main_db.Scan(ctx, cursor, pattern, 1000).Result()
// 			if err != nil {
// 				return nil, fmt.Errorf("error scanning keys for pattern %s: %v", pattern, err)
// 			}

// 			for _, key := range scannedKeys {
// 				allMatchingKeys = append(allMatchingKeys, key)
// 				keyToPatternIndex[key] = patternIndex
// 			}

// 			// Set cursor for the next scan iteration
// 			cursor = nextCursor
// 			if cursor == 0 {
// 				break
// 			}
// 		}
// 	}

// 	// If no matching keys are found, return an empty result
// 	if len(allMatchingKeys) == 0 {
// 		fmt.Println("No matching keys found in Redis.")
// 		return nil, nil
// 	}

// 	// Fetch the values for the matching keys
// 	values, err := db.main_db.MGet(ctx, allMatchingKeys...).Result()
// 	if err != nil {
// 		return nil, fmt.Errorf("error fetching values: %v", err)
// 	}

// 	return values, nil
// }

// func (db *DB) getRedisData() error {

// 	jsonData, err := db.traders_db.Get(ctx, strconv.Itoa(db.key)).Result()
// 	if err != nil {
// 		return fmt.Errorf("error retrieving data: %v", err)
// 	}

// 	var escapedData []string
// 	err = json.Unmarshal([]byte(jsonData), &escapedData)
// 	if err != nil {
// 		fmt.Printf("Error unmarshaling raw data: %v\n", err)
// 		return err
// 	}

// 	// Step 2: Iterate over the slice and unmarshal each escaped JSON string
// 	var parsedData [][]string
// 	for _, item := range escapedData {
// 		var innerData []string
// 		err := json.Unmarshal([]byte(item), &innerData)
// 		if err != nil {
// 			fmt.Printf("Error unmarshaling inner data: %v\n", err)
// 			return err
// 		}
// 		parsedData = append(parsedData, innerData)
// 	}

// 	// Print the cleaned data
// 	fmt.Println("Parsed Data:", parsedData)
// 	return nil
// }

// 2nd implementation where only one session id was fetched
// func (db *DB) fetchDataByPatterns(ctx context.Context, patterns [][]string) ([]interface{}, error) {
// 	var allResults []interface{}

// 	for _, patternGroup := range patterns {
// 		var allMatchingKeys []string

// 		for _, pattern := range patternGroup {
// 			cursor := uint64(0)
// 			fmt.Printf("Scanning for pattern: %s\n", patternGroup)

// 			for {
// 				scannedKeys, nextCursor, err := db.main_db.Scan(ctx, cursor, pattern, 1000).Result()
// 				if err != nil {
// 					return nil, fmt.Errorf("error scanning keys of Set '%d' for pattern %s: %v", db.Set, pattern, err)
// 				}

// 				allMatchingKeys = append(allMatchingKeys, scannedKeys...)

// 				cursor = nextCursor
// 				if cursor == 0 {
// 					break
// 				}
// 			}
// 		}

// 		if len(allMatchingKeys) == 0 {
// 			// Append an empty slice when no keys are found
// 			allResults = append(allResults, []interface{}{})
// 		} else {
// 			values, err := db.main_db.MGet(ctx, allMatchingKeys...).Result()
// 			if err != nil {
// 				return nil, fmt.Errorf("error fetching values for Set '%d': %v", db.Set, err)
// 			}

// 			allResults = append(allResults, values[0])
// 		}
// 	}

// 	return allResults, nil
// }
