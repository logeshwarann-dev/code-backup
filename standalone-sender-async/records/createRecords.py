import urllib.request
import zipfile
import os
import re
import csv
from collections import defaultdict

# URL of the zip file
url = 'https://www.bseindia.com/downloads1/Live_Masters_for_Equity_Derivatives.zip'
zip_file_path = 'Live_Masters_for_Equity_Derivatives.zip'
unzip_dir = 'equity_derivatives'
files = []

def download_file(url, file_path):
    """Downloads a file using urllib with a user-agent header."""
    try:
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
        }
        req = urllib.request.Request(url, headers=headers)
        with urllib.request.urlopen(req) as response:
            with open(file_path, 'wb') as out_file:
                out_file.write(response.read())
        print(f"File downloaded successfully: {file_path}")
    except Exception as e:
        print(f"Error downloading file: {e}")
        exit()

def unzip_file(zip_path, extract_to):
    """Unzips a file to the specified directory."""
    if not os.path.exists(extract_to):
        os.makedirs(extract_to)
    with zipfile.ZipFile(zip_path, 'r') as zip_ref:
        zip_ref.extractall(extract_to)
    print(f"Unzipped files to {extract_to}")

def delete_file(file_path):
    """Deletes the specified file."""
    if os.path.exists(file_path):
        os.remove(file_path)
        print(f"Deleted file: {file_path}")
    else:
        print(f"File not found: {file_path}")

def find_specific_files(directory):
    """Finds specific files based on patterns."""
    pattern1 = re.compile(r'^EQD_DP\d+$')
    pattern2 = re.compile(r'^BSE_EQD_CONTRACT_\d+\.csv$')

    for filename in os.listdir(directory):
        if pattern1.match(filename) or pattern2.match(filename):
            print(f"Matched file: {filename}")
            files.append(os.path.join(directory, filename))

# Step 1: Download the zip file
download_file(url, zip_file_path)

# Step 2: Unzip the downloaded file
unzip_file(zip_file_path, unzip_dir)
find_specific_files(unzip_dir)

# Step 3: Delete the zip file after extraction
delete_file(zip_file_path)

file1 = files[1]
file2 = files[0]

# Define necessary columns for file1 and file2
file1_data = []
with open(file1, mode='r') as f1:
    reader = csv.reader(f1)
    next(reader)  # Skip header
    for row in reader:
        file1_data.append({
            "InstrumentID": row[0],
            "Precision": row[1],
            "PreviousClosePrice": row[2],
            "LowerCircuit(%)": row[3],
            "UpperCircuit(%)": row[4],
            "LowerCircuitPrice": row[5],
            "UpperCircuitPrice": row[6]
        })

# Load file2 with necessary headers only
file2_data = []
with open(file2, mode='r') as f2:
    reader = csv.DictReader(f2)
    for row in reader:
        file2_data.append({
            "FinInstrmId": row["FinInstrmId"],
            "MinLot": row["MinLot"],
            "UnqPdctIdr": row["UnqPdctIdr"],
            "PartitionID": row["PartnIdr"]
        })

# Perform an inner join on InstrumentID and FinInstrmId
merged_data = []
for row1 in file1_data:
    for row2 in file2_data:
        if row1["InstrumentID"] == row2["FinInstrmId"]:
            merged_row = {
                "InstrumentID": row1["InstrumentID"],
                "Precision": row1["Precision"],
                "PreviousClosePrice": row1["PreviousClosePrice"],
                "LowerCircuit(%)": row1["LowerCircuit(%)"],
                "UpperCircuit(%)": row1["UpperCircuit(%)"],
                "LowerCircuitPrice": row1["LowerCircuitPrice"],
                "UpperCircuitPrice": row1["UpperCircuitPrice"],
                "MinLot": row2["MinLot"],
                "UnqPdctIdr": row2["UnqPdctIdr"],
                "PartitionID": row2["PartitionID"]
            }
            merged_data.append(merged_row)

# Group records by PartitionID
partitioned_data = defaultdict(list)
for row in merged_data:
    partitioned_data[row["PartitionID"]].append(row)

# Generate the output file
# Updated grouping logic for "Same ProductID" and "Diff ProductID"
output_file = "partitionedRecords.txt"
with open(output_file, mode='w') as f_out:
    # Write the header
    f_out.write("InstrumentID,Precision,PreviousClosePrice,LowerCircuit(%),UpperCircuit(%),LowerCircuitPrice,UpperCircuitPrice,MinLot,UnqPdctIdr,PartitionID\n")

    for partition_id, records in sorted(partitioned_data.items(), key=lambda x: int(x[0])):
        f_out.write(f"Partition {partition_id}:\n")

        # Group by ProductID (UnqPdctIdr)
        product_groups = defaultdict(list)
        for record in records:
            product_groups[record["UnqPdctIdr"]].append(record)

        # Same ProductID Group (5 records with the same UnqPdctIdr)
        f_out.write("Group: Same ProductID\n")
        same_product_group = []
        for product_id, product_records in product_groups.items():
            if len(product_records) >= 5:
                same_product_group = product_records[:5]
                break
        for record in same_product_group:
            f_out.write(','.join(str(record[col]) for col in [
                "InstrumentID", "Precision", "PreviousClosePrice", "LowerCircuit(%)",
                "UpperCircuit(%)", "LowerCircuitPrice", "UpperCircuitPrice",
                "MinLot", "UnqPdctIdr", "PartitionID"
            ]) + '\n')
        f_out.write("---------------------\n")

        # Diff ProductID Group (5 records with unique UnqPdctIdr, not in Same ProductID group)
        f_out.write("Group: Diff ProductID\n")
        diff_product_group = []
        used_product_ids = {record["UnqPdctIdr"] for record in same_product_group}
        unique_product_ids = set()

        for record in records:
            product_id = record["UnqPdctIdr"]
            if product_id not in used_product_ids and product_id not in unique_product_ids:
                diff_product_group.append(record)
                unique_product_ids.add(product_id)
                if len(diff_product_group) == 5:
                    break

        for record in diff_product_group:
            f_out.write(','.join(str(record[col]) for col in [
                "InstrumentID", "Precision", "PreviousClosePrice", "LowerCircuit(%)",
                "UpperCircuit(%)", "LowerCircuitPrice", "UpperCircuitPrice",
                "MinLot", "UnqPdctIdr", "PartitionID"
            ]) + '\n')
        f_out.write("---------------------\n")

print(f"Partitioned data successfully saved to {output_file}.")
