import redis # type: ignore
from config.Constants import *

def main():
    redis_client = connectToRedis()



def connectToRedis():
    try:
        redis_client = redis.Redis(host="localhost",port="6379",db=redisDbForConigFile)
        return redis_client
    except redis.ConnectionError as e:
        print(f"Redis connection error: {e}")
        return None
    except Exception as e:
        print(f"An error occurred: {e}")
        return None
    
# def staticMemberIds():
    LAB_ENV_IDS_WIN = [
        {
			"ID":       1034,
			"Password": "bse12345",
			"Status":   "A",
			"Traders": [
                Trader(tid=1001, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1002, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1003, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1014, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1017, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1019, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1020, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1021, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1022, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1023, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1024, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1025, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1026, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1027, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1028, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1029, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1030, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1031, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1032, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1033, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1034, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1035, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1036, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1037, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1038, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1039, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1040, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1041, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1042, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1043, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1044, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1045, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1046, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1047, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1048, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1049, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1050, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1051, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1052, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1053, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1054, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1055, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1056, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1057, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1058, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1059, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1060, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1061, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1062, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
                Trader(tid=1063, host_ip="192.168.87.29", port="11908", mid=1034, password="bse12345"),
            ]
        },
        {
		"ID":       1035,
		"Password": "bse12345",
		"Status":   "A",
		"Traders" : [],
        }
    ]