package utils

import (
	"github.com/go-redis/redis/v8"
)

type DB struct {
	main_db     *redis.Client
	traders_db  *redis.Client
	Set         int
	Addr        string
	session_ids []int
	key         int
}
