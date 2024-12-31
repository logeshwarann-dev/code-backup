package utils

import (
	"github.com/go-redis/redis/v8"
)

type DB struct {
	main_db     *redis.Client
	traders_db  *redis.Client
	Set         int
	Addr        string
	session_ids map[string]string
	key         int
}

type Config struct {
	CancelPercent            int    `json:"cancel_percent"`
	Duration                 int    `json:"duration"`
	EndTime                  string `json:"end_time"`
	EnvFile                  int    `json:"env_file"`
	Environment              int    `json:"environment"`
	Heartbeat                int    `json:"heartbeat"`
	MarketReplay             int    `json:"market_replay"`
	ModifyPercent            int    `json:"modify_percent"`
	RedisTimestampsBatchSize int    `json:"redis_tmstamps_batchsize"`
	StartTime                string `json:"start_time"`
	Throttle                 int    `json:"throttle"`
	Traders                  int    `json:"traders"`
	Instruments              []int  `json:"instruments"`
}
