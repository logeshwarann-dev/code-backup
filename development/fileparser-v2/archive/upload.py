import redis # type: ignore

def initialize_config_table(redis_client):
    """
    Create a single config entry in Redis as a hash with single values for each field.
    
    :param redis_client: Redis client instance.
    """
    try:
        # Define the configuration table with single values
        config_table = {
            "tempid": "10125",
            # "at": "30",
            "tif": "7",
            "cc": "CLIENT3"
        }

        # Store each key-value pair as a field in the Redis hash
        redis_client.hset("config", mapping=config_table)

        print("Config table successfully initialized as a Redis hash.")
    except Exception as e:
        print(f"Error initializing config table: {e}")


if __name__ == "__main__":
    try:
        # Connect to Redis
        redis_client = redis.Redis(host='localhost', port=6379, decode_responses=True)
        redis_client.ping()
        print("Connected to Redis")

        # Initialize the config table
        initialize_config_table(redis_client)
    except redis.ConnectionError as e:
        print(f"Redis connection error: {e}")
