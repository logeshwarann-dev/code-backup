from flask import Flask, jsonify, request
from modules.data.GenerateData import readFileAndGenerateData, GenerateDataforOrderPumping
from modules.redis.fetchDataFromRedis import FetchDataFromRedisAndTimeWarp
from modules.redis.redisConnection import redisConnection
from flask_cors import CORS 
from modules.userRequests.timewarpRequest import TimeWarp
from config.Constants import *
from modules.structs import *
from modules.userRequests.orderPumpingConfig import OrderPumpConfig
import threading

app = Flask(__name__)

# Enable CORS with specific configurations
CORS(app, resources={r"/*": {"origins": "*"}})

# Initialize Redis connection globally
redis_client = redisConnection()

if redis_client == None:
    print("redis Connection Fail")


@app.route('/api/v1/fileparser/set-market-replay-flag', methods=['POST'])
def set_market_replay_flag():
    req_body = request.get_json()
    MARKET_REPLAY = req_body.get('market_replay_flag') if req_body else None
    if not MARKET_REPLAY:
        return jsonify({"error": "The 'market_replay_flag' field is required."}), 400

    return jsonify({"message": f"Received Market Replay Flag: {MARKET_REPLAY}"}), 200

# ========ORDER PUMPING HANDLERS============
# STEP 1 (Optional)
@app.route('/api/v1/fileparser/order-pump/set-config', methods=['POST'])
def set_order_pumping_config():

    try: 
        order_pump_config_request_body = request.get_json()
        order_pump_config = OrderPumpConfig.from_json(order_pump_config_request_body)
        DEFAULT_TEMPLATEID = order_pump_config.user_template_id
        DEFAULT_ACCOUNT_TYPE = order_pump_config.user_account_type
        DEFAULT_TIME_INFORCE = order_pump_config.user_time_in_force
        DEFAULT_CLIENTCODE = order_pump_config.user_client_code
        DEFAULT_CLIENTORDERID = order_pump_config.user_client_order_id

        return jsonify({"message": "Order Pump config has been set successfully"}), 201
    except ValueError as e:
        return jsonify({"error": str(e)}), 400
    except Exception as e:
        return jsonify({"error": "An unexpected error occurred"}), 500

# STEP 2
#@app.route('/api/v1/fileparser/order-pump/redis/store-data-in-master', methods=['POST'])
def generate_order_pumping_data_and_store_in_master():

    if not redis_client:
        return jsonify({"error": "Redis connection failed"}), 500
     
    try:
        config_fields = get_config_fields(redis_client)
        redis_client.execute_command("SELECT 5")
        key_column_name = "FinInstrmId"
        value_columns_mapping = get_value_columns_mapping()
        GenerateDataforOrderPumping(redis_client, key_column_name, value_columns_mapping, config_fields)
        # return jsonify({"message": "Order Pumping Data has been loaded into Master Redis successfully"}), 201
        return "Order Pumping Data has been loaded into Master Redis successfully"
    except ValueError as e:
        # return jsonify({"error": str(e)}), 400
        return str(e)
    except Exception as e:
        # return jsonify({"error": "An unexpected error occurred"}), 500
        return "An unexpected error occurred"
    


# =========MARKET REPLAY HANDLERS============
# STEP 1 (Optional)
@app.route('/api/v1/fileparser/market-replay/set-timewarp-settings', methods=['POST'])
def set_timewarp_settings():
    try:
        # Parse and validate the request body
        timewarp_request_body = request.get_json()
        timewarp_settings = TimeWarp.from_json(timewarp_request_body)
        startTimeStr = timewarp_settings.start_time_str
        endTimeStr = timewarp_settings.end_time_str
        searchTraderId = timewarp_settings.search_trader_id
        searchPartitionId = timewarp_settings.search_partition_id
        searchProductId = timewarp_settings.search_product_id
        factorOfTimeWarping = timewarp_settings.factor_of_time_warping

        return jsonify({"message": "Timewarp setting done successfully"}), 201
    except ValueError as e:
        return jsonify({"error": str(e)}), 400
    except Exception as e:
        return jsonify({"error": "An unexpected error occurred"}), 500
 
# STEP 2
#@app.route('/api/v1/fileparser/market-replay/redis/store-data-in-master', methods=['POST'])
def generate_market_replay_data_and_store_in_master():
    """Endpoint to trigger data generation."""
    if not redis_client:
        return jsonify({"error": "Redis connection failed"}), 500

    try:
        redis_client.execute_command("SELECT 5")
        thread = threading.Thread(target=readFileAndGenerateData, args=(redis_client,), daemon=True)
        thread.start()
        # readFileAndGenerateData(redis_client)
        #return jsonify({"message": "Data generated successfully"}), 200
        return "File Parser started loading data into Master Redis"
    except Exception as e:
        return str(e)

# STEP 3
@app.route('/api/v1/fileparser/market-replay/redis/timewarp', methods=['POST'])
def fetch_and_timewarp_data():
    """Endpoint to fetch data and perform time warp."""
    try:
        redis_client.execute_command("SELECT 5")
        FetchDataFromRedisAndTimeWarp(redis_client)
        return jsonify({"message": "Data fetched and time-warp operation completed"}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500

# ===============
@app.route('/api/v1/fileparser/process-file', methods=['POST'])
def process_file():
    if not redis_client:
        return jsonify({"error": "Redis connection failed"}), 500
    
    if MARKET_REPLAY:
        # req_body = request.get_json()
        # filePath = req_body.get('file_path') if req_body else None
        response = generate_market_replay_data_and_store_in_master()
        return jsonify({"message": response}), 200

    else:
        # req_body = request.get_json()
        # FILEPATH_ORDERPUMPING = req_body.get('file_path') if req_body else None
        response = generate_order_pumping_data_and_store_in_master()
        return jsonify({"message": response}), 200

    






# ==========HEALTH CHECK=======
@app.route('/api/v1/fileparser/health-check', methods=['POST'])
def health_status():
    """Health check endpoint."""
    if redis_client:
        return jsonify({"status": "File Parser pod is Active"}), 200
    else:
        return jsonify({"status": "Redis connection failed"}), 500



# ==========Archive===========

@app.route('/api/v1/file/upload', methods=['POST'])
def store_history_file():
    # Handle file upload
    if 'file' not in request.files:
        print(f"Requests: {request.files}")
        return {"errors": ["File not provided"]}, 422
    
    uploaded_file = request.files['file']
    print(f"Filenaem: {uploaded_file.filename}")
    
    if uploaded_file.filename != "":
        try:
            uploaded_file.save(f'./static/HistoryFile.csv')
            return {"filename": uploaded_file.filename, "message": "File has been successfully uploaded"}, 201
        except Exception as err:
            return {"errors": [str(err)]}, 422
    return {"errors": ["missing filename"]}, 422




if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
