import redis # type: ignore
from config.Constants import *
import os

def redisConnection():
    # print("Environment variables:", os.environ)
    redisAddr = os.environ['master-redis']
    # redisAddr = "34.229.19.59"

    try:
        redis_client = redis.Redis(host=redisAddr, port=6379, db=redisDb)
        redis_client.ping()

        print("Successfully connected to Redis")
        return redis_client
    except redis.ConnectionError as e:
        print(f"Redis connection error: {e}")
        return None
    except Exception as e:
        print(f"An error occurred: {e}")
        return None