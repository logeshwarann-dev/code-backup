import redis # type: ignore
from config.Constants import *
import os

def redisConnection():

    redisAddr = os.environ['master-redis']
    # redisAddr = "localhost"


    try:
        redis_client = redis.Redis(host=redisAddr, port=6379, db=redisDbForData)
        redis_client.ping()
        
        count_redis_client = redis.Redis(host=redisAddr, port=6379, db=redisDbForCount)
        count_redis_client.ping()

        print("Successfully connected to Redis")
        return redis_client, count_redis_client
    except redis.ConnectionError as e:
        print(f"Redis connection error: {e}")
        return None, None
    except Exception as e:
        print(f"An error occurred: {e}")
        return None, None
