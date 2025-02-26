package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func RedisConnection() error {
	ctx := context.Background()

	//------------------------------------
	masterRedisIP := os.Getenv("master-redis")
	masterRedisURL := fmt.Sprintf("%v:6379", masterRedisIP)

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

	if err := SetConfig(mainRdb); err != nil {
		fmt.Println("Error in HSET of main db:", err)
		return err
	}

	if err := SetConstants(mainRdb); err != nil {
		return err
	}

	if err := ScaleRedisAndSenderPods(len(MEMBER_SET)); err != nil {
		return fmt.Errorf("unable to scale redis and sender pods: %v", err)
	}

	redisPods := GetRedisPodList(len(MEMBER_SET))

	constructTradersDB(redisPods)

	// var wg sync.WaitGroup
	for index := range len(MEMBER_SET) {
		// wg.Add(1)
		database := &Traders_DB[index]
		go func(db *DB) {
			// defer wg.Done()

			traderClient := redis.NewClient(&redis.Options{
				Addr: db.Addr,
				//Addr: "localhost:6379",
				DB: 0,
			})

			for retryConn := range 5 {
				if _, err := traderClient.Ping(ctx).Result(); err != nil {
					fmt.Printf("Failed to connect to trader Redis: %v\n", err)
					fmt.Println("Retrying Redis Pod Connection: Attempt-", retryConn)
					time.Sleep(5 * time.Second)
					continue
				}
				// fmt.Println("Redis Pod Connection Successful!")
				break
			}

			db.main_db = mainRdb
			db.traders_db = traderClient

			fmt.Println("REDIS CONNECTED SUCCESSFULLY FOR TRADER SET:", db.Set)

			err := db.SetDataToSecondaryRedis(index)
			if err != nil {
				fmt.Println("Error in Fetch from Redis", err)
			}

		}(database)
	}

	// wg.Wait()
	return nil
}
