import redis

def flush_all_databases():
    try:
        # Connect to the Redis server
        r = redis.StrictRedis(host='54.237.6.186', port=6379, decode_responses=True)

        # Ping the server to ensure the connection is successful
        if r.ping():
            print("Connected to Redis server.")

        # Flush all databases
        r.flushall()
        print("All databases in Redis have been successfully cleared.")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    flush_all_databases()
