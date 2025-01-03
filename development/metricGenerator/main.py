from MemberSetOrderPumping import *  
from constants import *  
from metrics import *    
import json
import redis #type: ignore
from MemberSetOrderPumping import *

def main():
    # Store member set into Redis
    redis_Client = redis.Redis(host='localhost', port=6379, db=MetricsDB)
    if(redis_Client == None):
        print("Redis Connection Failed")
        return

    if MARKET_REPLAY == False:
        set_member_ids_for_order_pumping(redis_Client)
        return

    TotalUniqueSessionIds = 0

    # Start metrics calculation
    instrument_id_count, session_id_data, session_id_count, buy_sell_counts, order_type_counts, hourly_counts = start_member_set_and_metrics_calculation(csv_files)

    # Convert `defaultdict` to `dict` for JSON serialization
    order_type_counts = {key: dict(value) for key, value in order_type_counts.items()}
    buy_sell_counts = {key: dict(value) for key, value in buy_sell_counts.items()}

    # Calculate derived metrics
    TotalUniqueSessionIds = len(session_id_count)
    top_10_session_counts = sorted(session_id_count.items(), key=lambda x: x[1], reverse=True)[:10]
    top_5_instrument_counts = sorted(instrument_id_count.items(), key=lambda x: x[1], reverse=True)[:5]
    top_hour_group = max(hourly_counts, key=hourly_counts.get)
    lowest_hour_group = min(hourly_counts, key=hourly_counts.get)

    # Prepare metrics for storage
    metrics = {
        "top_5_instrument_ids": str(top_5_instrument_counts),
        "top_10_session_ids": str(top_10_session_counts),
        "total_unique_session_ids": TotalUniqueSessionIds,
        "buy_sell_counts": json.dumps(buy_sell_counts),  # Convert to JSON for storage
        "order_type_counts": json.dumps(order_type_counts),  # Convert to JSON for storage
        "top_hour_group": top_hour_group,
        "top_hour_group_orders": hourly_counts[top_hour_group],
        "lowest_hour_group": lowest_hour_group,
        "lowest_hour_group_orders": hourly_counts[lowest_hour_group]
    }

    store_member_set_to_redis(session_id_data,redis_Client)

    # Store metrics into Redis
    store_metrics_to_redis(metrics,redis_Client)


if __name__ == "__main__":
    main()
