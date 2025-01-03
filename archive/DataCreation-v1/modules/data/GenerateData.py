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

def readFileAndGenerateData(redis_client, count_redis_client):
    total_entries = 0
    failed = 0
    count = defaultdict()
    tempMap = defaultdict(list)

    try:
        with open(filePath, mode='r', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            
            for row in reader:
                for key, value in row.items():
                    if value is None or value.strip() == "":
                        row[key] = ""
                
                if(row["Transaction Timestamp"] == ""):
                    continue
                
                typeOfOrder = SingleLegLeanOrder
                data, traderId, productId = generatedata(typeOfOrder, row)

                if data is None or traderId is None or productId is None:
                    failed += 1
                    continue

                key = f"{int(datetime.strptime(row['Transaction Timestamp'], '%Y-%m-%d %H:%M:%S').replace(tzinfo=timezone.utc).timestamp())}:{traderId}:{partitionId}:{productId}"
                
                tempMap[key].append(data.toCommaSeparated())

                if key not in count:
                    count[key] = 0

                count[key] += 1

                total_entries += 1

    except Exception as e:
        print(f"Error reading file: {e}")
        return

    for key, value in tempMap.items():
        d = json.dumps(value)
        redis_client.set(key, d)

    for key,value in count.items():
        count_redis_client.set(key,value)

    print(f"Total Entries Done = {total_entries} and Total Failed Entries = {failed}")
    count_redis_client.set("total_entries", total_entries)


# import io
# import pandas as pd # type: ignore
# import time
# from ParseData import *
# from datetime import datetime, timedelta
# import json
# import random
# from collections import defaultdict
# from Constants import *

# def generatedata(typeOfOrder,row):
#     switch = {
#         SingleLegStandardOrder: lambda: generateNewSingleLegOrder(0, row), 
#         SingleLegLeanOrder: lambda: singleLegLeanOrder(0, row),  
#     }

#     handler = switch.get(typeOfOrder, lambda: (None,None,None))

#     try:
#         result = handler() 
#         if typeOfOrder == SingleLegStandardOrder:
#             data,traderId,productId = result
#             return data,traderId,productId
#         elif typeOfOrder == SingleLegLeanOrder:
#             data,traderId,productId = result
#             return data,traderId,productId
#         else:
#             raise ValueError("Invalid order type")

#     except Exception as e:
#         print(f"Error while generating data: {e}")
#         return None,None,None



# # def readFileAndGenerateData(redis_client,count_redis_client):
# #     df = pd.read_excel(filePath,sheet_name = sheetName)
    
# #     df.fillna(0, inplace=True)

# #     total_entries = 0
# #     failed = 0

# #     tempMap = defaultdict(list)

# #     # for _ in range(total_time):
# #     for index,row in df.iterrows():
# #         typeOfOrder = SingleLegLeanOrder
# #         data,traderId,productId = generatedata(typeOfOrder,row)

# #         if data == None or traderId == None or productId == None:
# #             failed+=1
# #             continue
        
# #         key = str(int(row["Transaction Timestamp"].timestamp()))
# #         # print(row["Transaction Timestamp"])
# #         # print(key)
# #         key += ":"
# #         key += str(traderId)
# #         key += ":"
# #         key += str(partitionId)
# #         key += ":"
# #         key += str(productId)

# #         tempMap[key].append(data.to_dict())
        
# #         total_entries += 1

# #         # print(total_entries)

# #         # if time.time() - current_epoch_second > 1:
# #         #     break  

            
# #     for key,value in tempMap.items():    
# #         d = json.dumps(value)
# #         redis_client.set(key,d)

# #     # count_redis_client.set(str(current_epoch_second), count)

# #         # total_entries += count

# #         # elapsed_time = time.time() - current_epoch_second
# #         # sleep_time = 1 - elapsed_time  # Remaining time to sleep

# #         # if sleep_time > 0:
# #         #     time.sleep(sleep_time)

# #     print(f"Total Entries Done = {total_entries} and Total Failed Entries = {failed}")

# #     count_redis_client.set("total_entries", total_entries)





# def readFileAndGenerateData(redis_client,count_redis_client):
#     df = pd.read_excel(filePath,sheet_name = sheetName)
    
#     df.fillna(0, inplace=True)

#     total_entries = 0
#     failed = 0

#     tempMap = defaultdict(lambda: [])

#     # for _ in range(total_time):
#     for index,row in df.iterrows():
#         typeOfOrder = SingleLegLeanOrder
#         data,traderId,productId = generatedata(typeOfOrder,row)

#         if data == None or traderId == None or productId == None:
#             failed+=1
#             continue
        
#         key = str(int(row["Transaction Timestamp"].timestamp()))
#         # print(row["Transaction Timestamp"])
#         # print(key)
#         key += ":"
#         key += str(traderId)
#         key += ":"
#         key += str(partitionId)
#         key += ":"
#         key += str(productId)

#         tempMap[key].append(data.toCommaSeparated())
        
#         total_entries += 1

#         # print(total_entries)

#         # if time.time() - current_epoch_second > 1:
#         #     break  

            
#     for key,value in tempMap.items():    
#         d = json.dumps(value)
#         redis_client.set(key,d)

#     # count_redis_client.set(str(current_epoch_second), count)

#         # total_entries += count

#         # elapsed_time = time.time() - current_epoch_second
#         # sleep_time = 1 - elapsed_time  # Remaining time to sleep

#         # if sleep_time > 0:
#         #     time.sleep(sleep_time)

#     print(f"Total Entries Done = {total_entries} and Total Failed Entries = {failed}")

#     count_redis_client.set("total_entries", total_entries)
