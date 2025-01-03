from concurrent.futures import ThreadPoolExecutor
import csv
from datetime import datetime, time
from collections import defaultdict
import json

# Function to determine the numeric time group mapping
def get_time_group(timestamp):
    time_obj = timestamp.time()
    if time(9, 15) <= time_obj <= time(10, 15):
        return 0
    elif time(10, 15) < time_obj <= time(11, 15):
        return 1
    elif time(11, 15) < time_obj <= time(12, 15):
        return 2
    elif time(12, 15) < time_obj <= time(14, 15):
        return 3
    elif time(14, 15) < time_obj <= time(15, 30):
        return 4
    else:
        return None

# Function to process a single file and return the counts
def metrics(file_path):
    count_for_instrument_id = {}
    session_id_data = {}
    session_id_count = defaultdict(int)
    buy_sell_counts = defaultdict(lambda: {"BUY": 0, "SELL": 0})
    order_type_counts = defaultdict(lambda: {"Order Entry": 0, "Order Modification": 0, "Order Cancellation": 0})
    time_groups = range(5)  # Time groups 0, 1, 2, 3, 4
    hourly_counts = defaultdict(int, {group: 0 for group in time_groups})

    try:
        with open(file_path, newline='', encoding='utf-8') as csvfile:
            reader = csv.DictReader(csvfile)
            for row in reader:
                # Count Instrument Id
                instrument_id = int(row["Instrument ID"])
                if instrument_id not in count_for_instrument_id:
                    count_for_instrument_id[instrument_id] = 0
                count_for_instrument_id[instrument_id] += 1

                # Calculate and count Session Id
                member_id = int(row["Trading Member Id"])
                trader_id = int(row["Trader Id"])
                sid = (member_id * 100000) + trader_id
                if sid not in session_id_data:
                    session_id_data[sid] = {
                        "TID": trader_id,
                        "MID": member_id,
                        "session_id": sid
                    }
                session_id_count[sid] += 1

                # Parse the timestamp and group by numeric time intervals
                timestamp_str = row["Transaction Timestamp"]
                timestamp = datetime.strptime(timestamp_str, "%Y-%m-%d %H:%M:%S")
                time_group = get_time_group(timestamp)

                if time_group is not None:
                    # Count orders based on Buy/Sell type
                    order_buy_sell = "BUY" if row["Buy Sell Type"] == "1" else "SELL"
                    buy_sell_counts[time_group][order_buy_sell] += 1

                    # Count orders based on Order Type
                    order_type_str = {
                        "1": "Order Entry",
                        "2": "Order Modification"
                    }.get(row["Transaction Type"], "Order Cancellation")
                    order_type_counts[time_group][order_type_str] += 1

                    # Increment hourly count for the time group
                    hourly_counts[time_group] += 1

    except Exception as e:
        print(f"Error processing {file_path}: {e}")
    return count_for_instrument_id, session_id_data, session_id_count, buy_sell_counts, order_type_counts, hourly_counts

# Function to aggregate metrics across all files
def start_member_set_and_metrics_calculation(csv_files):
    aggregated_instrument_counts = {}
    aggregated_session_data = {}
    aggregated_session_counts = defaultdict(int)
    aggregated_buy_sell_counts = defaultdict(lambda: {"BUY": 0, "SELL": 0})
    aggregated_order_type_counts = defaultdict(lambda: {"Order Entry": 0, "Order Modification": 0, "Order Cancellation": 0})
    time_groups = range(5)  # Time groups 0, 1, 2, 3, 4
    aggregated_hourly_counts = defaultdict(int, {group: 0 for group in time_groups})

    with ThreadPoolExecutor(max_workers=11) as executor:
        futures = [executor.submit(metrics, file_path) for file_path in csv_files]

        for future in futures:
            try:
                (file_instrument_counts, file_session_data, file_session_counts, 
                 file_buy_sell_counts, file_order_type_counts, file_hourly_counts) = future.result()

                # Aggregate Instrument Id counts
                for instrument_id, count in file_instrument_counts.items():
                    aggregated_instrument_counts[instrument_id] = aggregated_instrument_counts.get(instrument_id, 0) + count

                # Aggregate Session Id data and counts
                aggregated_session_data.update(file_session_data)
                for sid, count in file_session_counts.items():
                    aggregated_session_counts[sid] += count

                # Aggregate Buy/Sell counts by time group
                for time_group, counts in file_buy_sell_counts.items():
                    aggregated_buy_sell_counts[time_group]["BUY"] += counts["BUY"]
                    aggregated_buy_sell_counts[time_group]["SELL"] += counts["SELL"]

                # Aggregate Order Type counts by time group
                for time_group, counts in file_order_type_counts.items():
                    aggregated_order_type_counts[time_group]["Order Entry"] += counts["Order Entry"]
                    aggregated_order_type_counts[time_group]["Order Modification"] += counts["Order Modification"]
                    aggregated_order_type_counts[time_group]["Order Cancellation"] += counts["Order Cancellation"]

                # Aggregate Hourly counts
                for time_group, count in file_hourly_counts.items():
                    aggregated_hourly_counts[time_group] += count

            except Exception as e:
                print(f"Error aggregating results: {e}")

    return (aggregated_instrument_counts, aggregated_session_data, aggregated_session_counts, 
            aggregated_buy_sell_counts, aggregated_order_type_counts, aggregated_hourly_counts)

# Function to store metrics in Redis
def store_metrics_to_redis(metrics,r):
    try: 
        r.execute_command("SELECT 2")
        for key, value in metrics.items():
            r.hset("metrics", key, value)
        print("Metrics successfully stored in Redis.")
    except Exception as e:
        print(f"Error storing metrics to Redis: {e}")


def store_member_set_to_redis(session_id,r):
    try:
        for key, value in session_id.items():
            r.hset("member_set", key, json.dumps(value))
        print("Member set successfully stored in Redis.")
    except Exception as e:
        print(f"Error storing member set to Redis: {e}")