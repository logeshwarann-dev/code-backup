from flask import Flask, jsonify, request
from flask_cors import CORS 
import redis 
import os
from concurrent.futures import ThreadPoolExecutor
from MemeberSet import *
from constants import *
from constants import *  
from metrics import *    
import json

app = Flask(__name__)

# Enable CORS with specific configurations
CORS(app, resources={r"/*": {"origins": "*"}})



@app.route('/api/v1/metric-gen/store-metrics-in-redis', methods=['POST'])
def generate_metrics():
    unique_ids = {}

    # Use ThreadPoolExecutor to process files in parallel
    with ThreadPoolExecutor(max_workers=11) as executor:
        futures = [executor.submit(process_file, file_path) for file_path in csv_files]
        for future in futures:
            unique_ids.update(future.result())

    # Store the combined unique IDs into Redis
    store_to_redis(unique_ids)
    return jsonify({"message": "Metrics are getting stored in Master Redis"}), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5001, debug=True)