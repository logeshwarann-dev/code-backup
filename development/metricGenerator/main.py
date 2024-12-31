from concurrent.futures import ThreadPoolExecutor
from MemeberSet import *
from constants import *

def main():
    unique_ids = {}

    # Use ThreadPoolExecutor to process files in parallel
    with ThreadPoolExecutor(max_workers=11) as executor:
        futures = [executor.submit(process_file, file_path) for file_path in csv_files]
        for future in futures:
            unique_ids.update(future.result())

    # Store the combined unique IDs into Redis
    store_to_redis(unique_ids)

if __name__ == "__main__":
    main()
