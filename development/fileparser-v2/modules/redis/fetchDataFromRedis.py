from config.Constants import *
from modules.structs import RedisData
import json
from modules.timeWarping import timeWarp
from collections import defaultdict
from datetime import datetime, timezone
import os


def FetchDataFromRedisAndTimeWarp(rdb):

    startTime = int(datetime.strptime(startTimeStr, "%d-%m-%Y %H:%M:%S").replace(tzinfo=timezone.utc).timestamp())
    print(datetime.strptime(startTimeStr, "%d-%m-%Y %H:%M:%S"))
    print(startTime)
    tempStartTime = startTime
    reference = startTime

    count = {}
    newData = defaultdict(list)
    totalbefore = 0
    num = 0
    totalafter = 0

    for t in range(0, duration):
        data_items = fetch_data_by_pattern(rdb, f"{tempStartTime}:{searchTraderId}:{searchPartitionId}:{searchProductId}")

        if len(data_items) > 0:
            newtime = timeWarp(tempStartTime, reference, factorOfTimeWarping)

            if newtime not in count:
                count[newtime] = 0

            for items in data_items:
                num += 1
                newData[newtime].append(items)
            count[newtime] += len(data_items)
            totalbefore += len(data_items)

        tempStartTime += 1

    writeToRedis(newData,rdb)
    print(f"TOTAL BEFORE: {totalbefore}")
    numberOfEntries = 0
    for a in count:
        totalafter += count[a]
        numberOfEntries += 1
    print(count)

    print(f"TOTAL AFTER : {totalafter}")
    print(f"Number of Entries generated : {numberOfEntries}")

def writeToRedis(new_data,redisConn):
    redisConn.execute_command("SELECT 7")
    tempMap = defaultdict(list)
    
    print(len(new_data))

    for k, value_list in new_data.items():
        try:
            for obj in value_list:
                if isinstance(obj, RedisData):
                    newKey = str(k)
                    newKey += ":" + str(obj.trader_id) + ":" + str(obj.partition_id) + ":" + str(obj.product_id)
                    tempMap[newKey].append(obj.toCommaSeparated())
        except Exception as e:
            print(f"Error processing key {k}: {e}")
    
    for key, value in tempMap.items():
        try:
            serialized_value = json.dumps(value)  # Store as a JSON array
            redisConn.set(key, serialized_value)
        except Exception as e:
            print(f"Error writing key {key} to Redis: {e}")



def fetch_data_by_pattern(rdb, pattern):
    cursor = 0
    data_items = []
    batch_size = 100000000

    while True:
        cursor, keys = rdb.scan(cursor, match=pattern, count=batch_size)

        if not keys:
            if cursor == 0:
                break
            continue

        values = rdb.mget(keys)

        for i, value in enumerate(values):
            if value is None:
                continue

            try:
                str_value = value.decode("utf-8")
                data_items.extend(process_value(keys[i].decode("utf-8"), str_value))
            except Exception as e:
                print(f"Error decoding value for key {keys[i].decode('utf-8')}: {e}")
                continue

        if cursor == 0:
            break

    return data_items


def process_value(key, value):
    parts = key.split(":")
    if len(parts) < 4:
        raise ValueError("Invalid key format")

    timestamp = int(parts[0])
    trader_id = int(parts[1])
    partition_id = int(parts[2])
    product_id = int(parts[3])

    data_items = []

    try:
        data_list = json.loads(value)
        for item in data_list:
            arr = item.split(",")
            if len(arr) < 10:
                raise ValueError(f"Invalid entry in Redis data: {item}")
            data_item = RedisData(
                template_id=int(arr[0]),
                instid=int(arr[1]),
                price=float(arr[2]),
                order_qty=int(arr[3]),
                maxshow=int(arr[4]),
                acc_type=int(arr[5]),
                time_in_force=int(arr[6]),
                clientCode=arr[7],
                MsgSeqNum=int(arr[8]),
                transactionType=arr[9],
                OrderNumber= arr[10],
                TransactionTimeStamp=int(arr[11]),
                TradingProductId=int(arr[12]),
                memberID=int(arr[13]),
                actualTraderID=int(arr[14]),                
                PartitionID=int(arr[15]),
                trader_id=trader_id,
                partition_id=partition_id,
                product_id=product_id,
                timestamp=timestamp,
                
            )
            data_items.append(data_item)

    except Exception as e:
        print(f"Error processing key {key}: {e}")

    return data_items


