from flask import Flask, jsonify, request
from modules.data.GenerateData import readFileAndGenerateData
from modules.redis.fetchDataFromRedis import FetchDataFromRedisAndTimeWarp
from modules.redis.redisConnection import redisConnection
from flask_cors import CORS

app = Flask(__name__)

# Enable CORS with specific configurations
CORS(app, resources={r"/*": {"origins": "*"}})

# Initialize Redis connection globally
redis_client,count_redis_client = redisConnection()

if redis_client == None or count_redis_client == None:
    print("redis Connection Fail")


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


@app.route('/api/v1/fileparser/redis/store-data-in-master', methods=['POST'])
def generate_data_route():
    """Endpoint to trigger data generation."""
    if not redis_client or not count_redis_client:
        return jsonify({"error": "Redis connection failed"}), 500
    
    try:
        readFileAndGenerateData(redis_client,count_redis_client)
        return jsonify({"message": "Data generated successfully"}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500



@app.route('/api/v1/fileparser/redis/timewarp', methods=['POST'])
def fetch_and_timewarp_route():
    """Endpoint to fetch data and perform time warp."""
    try:
        FetchDataFromRedisAndTimeWarp()
        return jsonify({"message": "Data fetched and time-warp operation completed"}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/api/v1/fileparser/status', methods=['GET'])
def status():
    """Health check endpoint."""
    if redis_client and count_redis_client:
        return jsonify({"status": "OK"}), 200
    else:
        return jsonify({"status": "Redis connection failed"}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
