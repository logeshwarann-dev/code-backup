from GenerateData import readFileAndGenerateData
from redisConnection import redisConnection
from fetchDataFromRedis import FetchDataFromRedisAndTimeWarp

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
