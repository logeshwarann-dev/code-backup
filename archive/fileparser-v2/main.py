from modules.data.GenerateData import readFileAndGenerateData, GenerateDataforOrderPumping
from modules.redis.redisConnection import redisConnection
from modules.redis.fetchDataFromRedis import FetchDataFromRedisAndTimeWarp
from config.Constants import *
from modules.structs import *
from concurrent.futures import ThreadPoolExecutor

def main():
    # Establish connection to Redis (using db=5 for application logic)
    redis_client = redisConnection()

    if redis_client is None:
        print("Redis connection failed")
        return

    # Check if running in Market Replay mode
    if MARKET_REPLAY:

        redis_client.execute_command("SELECT 5")

        readFileAndGenerateData(redis_client) # parameter = file path is the input

        FetchDataFromRedisAndTimeWarp(redis_client) # refer to parameters in constants.py

    else:
        # Retrieve configuration fields from Redis database 0
        config_fields = get_config_fields(redis_client) 

        redis_client.execute_command("SELECT 5")

        key_column_name = "FinInstrmId"
        
        value_columns_mapping = get_value_columns_mapping()

        GenerateDataforOrderPumping(redis_client, key_column_name, value_columns_mapping, config_fields)

    print("Done")


if __name__ == "__main__":
    main()
