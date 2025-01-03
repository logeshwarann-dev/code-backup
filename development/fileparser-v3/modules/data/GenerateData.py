import csv
import time
from datetime import datetime, timezone
import json
from collections import defaultdict
from modules.data.ParseData import *
from config.Constants import *

def generatedata(typeOfOrder, row):
    switch = {
        SingleLegStandardOrder: lambda: generateNewSingleLegOrder(0, row),
        SingleLegLeanOrder: lambda: singleLegLeanOrder(0, row),
    }

    handler = switch.get(typeOfOrder, lambda: (None, None, None))

    try:
        result = handler()
        if typeOfOrder == SingleLegStandardOrder:
            data, traderId, productId = result
            return data, traderId, productId
        elif typeOfOrder == SingleLegLeanOrder:
            data, traderId, productId = result
            return data, traderId, productId
        else:
            raise ValueError("Invalid order type")

    except Exception as e:
        print(f"Error while generating data: {e}")
        return None, None, None

def readFileAndGenerateData(redis_client):
    total_entries = 0
    failed = 0
    count = defaultdict()
    tempMap = defaultdict(list)

    startTime = -1
    currTime = -1
    tempCount = 0
    prevTime = -1

    try:
        with open(filePath, mode='r', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            for row in reader:
                for key, value in row.items():
                    if value is None or value.strip() == "":
                        row[key] = ""
                
                if(row["Transaction Timestamp"] == ""):
                    continue
                
                if startTime == -1:
                    startTime = int(datetime.strptime(startTimeStr, '%d-%m-%Y %H:%M:%S').replace(tzinfo=timezone.utc).timestamp())
                    prevTime = startTime
                typeOfOrder = SingleLegLeanOrder
                data, traderId, productId = generatedata(typeOfOrder, row)

                if data is None or traderId is None or productId is None:
                    failed += 1
                    continue

                key = f"{int(datetime.strptime(row['Transaction Timestamp'], '%Y-%m-%d %H:%M:%S').replace(tzinfo=timezone.utc).timestamp())}:{traderId}:{partitionId}:{productId}"
                currTime = int(datetime.strptime(row['Transaction Timestamp'], '%Y-%m-%d %H:%M:%S').replace(tzinfo=timezone.utc).timestamp())
                
                tempMap[key].append(data.toCommaSeparated())

                if key not in count:
                    count[key] = 0

                count[key] += 1

                total_entries += 1
                tempCount += 1
                if prevTime != currTime:
                    prevTime = currTime
                    print(f"change in time {currTime}")
                    print(f"Count = {tempCount}")
                    if(tempCount == COUNT or tempCount >= (0.9*COUNT)):
                        tempCount = 0
                        timeDiff = currTime - startTime
                        startTime = currTime
                        timeDiff /= 2
                        
                        for key, value in tempMap.items():
                            d = json.dumps(value)
                            redis_client.set(key, d)
                        tempMap.clear() 

                        print(f"Count = {total_entries} sleeping for {timeDiff} seconds")
                        time.sleep(timeDiff)

            if len(tempMap) > 0:
                for key, value in tempMap.items():
                    d = json.dumps(value)
                    redis_client.set(key, d)
                tempMap.clear()

    except Exception as e:
        print(f"Error reading file: {e}")
        return

    print(f"Total Entries Done = {total_entries} and Total Failed Entries = {failed}")


def GenerateDataforOrderPumping(redis_client, key_column_name, value_columns_mapping, config_fields):
    """
    Generate and store order pumping data in Redis under a single key 'ORDER_PUMPING'.

    :param redis_client: Redis client connection (DB 5).
    :param csv_file_path: Path to the CSV file.
    :param key_column_name: Column in the CSV to use as the key for each entry.
    :param value_columns_mapping: Mapping of CSV columns to Redis field names.
    :param config_fields: Default configuration fields to fill missing values.
    """
    try:
        # Read the CSV file and process data
        with open(FILEPATH_ORDERPUMPING, mode="r") as csv_file:
            reader = csv.DictReader(csv_file)
            order_pumping_data = {}

            for row in reader:
                # Extract the instrument ID to use as the Redis hash field name
                instr_id = row[key_column_name]

                # Prepare the value for this instrument ID
                instrument_data = {}
                for csv_col, redis_field in value_columns_mapping.items():
                    # Use value from CSV if present, otherwise use config fields
                    instrument_data[redis_field] = row.get(csv_col, config_fields.get(redis_field, "N/A"))

                # Serialize the instrument data as JSON
                order_pumping_data[instr_id] = json.dumps(instrument_data)

            # Store all instrument data under a single Redis hash key 'ORDER_PUMPING'
            redis_client.hset("ORDER_PUMPING", mapping=order_pumping_data)
            print("All data successfully stored under 'ORDER_PUMPING' key.")

    except Exception as e:
        print(f"Error generating data for order pumping: {e}")
        return
