import csv
from collections import Counter

# File path to the CSV
file_path = r"static/BSE_EQD_CONTRACT_20122024.csv"

# Initialize counters
company_counter = Counter()
option_type_counter = Counter()

# Read the CSV and process rows
with open(file_path, mode='r') as file:
    reader = csv.DictReader(file)
    for row in reader:
        # Count orders per company
        company = row['TckrSymb']
        if company:
            company_counter[company] += 1

        # Count option type distribution
        option_type = row['OptnTp']
        if option_type:
            option_type_counter[option_type] += 1


# Display results
print("Orders per Company:")
for company, count in company_counter.most_common(10):  # Top 10 companies
    print(f"{company}: {count} orders")

print("\nOption Type Distribution:")
for option_type, count in option_type_counter.items():
    print(f"{option_type}: {count} orders")
