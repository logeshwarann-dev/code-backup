import csv
import redis  # type: ignore
import json
import os
from constants import *

# Function to read a single CSV file and extract unique session IDs
def process_file(file_path):
    unique_ids = {}
    try:
        with open(file_path, newline='', encoding='utf-8') as csvfile:
            reader = csv.DictReader(csvfile)
            for row in reader:
                member_id = int(row["Trading Member Id"])
                trader_id = int(row["Trader Id"])
                session_id = (member_id * 100000) + trader_id

                if session_id not in unique_ids:
                    unique_ids[session_id] = {
                        "TID": trader_id,
                        "MID": member_id,
                        "session_id": session_id
                    }

        print(f"Successfully processed: {file_path}")
    except Exception as e:
        print(f"Error processing {file_path}: {e}")
    return unique_ids



# Function to store all unique IDs in a single Redis document
def store_to_redis(unique_ids):
    redisAddr = os.environ['master-redis']
    try:
        redis_connection = redis.Redis(host=redisAddr, port=6379, db = ConfigDB)
        redis_connection.hset("member_set", mapping={
            str(session_id): json.dumps(data) for session_id, data in unique_ids.items()
        })
        print("All unique IDs successfully stored in Redis under 'member_set'.")
    except Exception as e:
        print(f"Error storing data to Redis: {e}")

