from modules.data.GenerateData import readFileAndGenerateData
from modules.redis.redisConnection import redisConnection
from modules.redis.fetchDataFromRedis import FetchDataFromRedisAndTimeWarp

def main():
    redis_client,count_redis_client = redisConnection()

    if redis_client == None or count_redis_client == None:
        print("redis Connection Fail")
        return

    readFileAndGenerateData(redis_client,count_redis_client)

    FetchDataFromRedisAndTimeWarp()

    print("Done")
    

if __name__ == "__main__":
    main()
