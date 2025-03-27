import argparse
import sqlite3
import pandas as pd
import datetime
import json
import sys
import os
import random

parser = argparse.ArgumentParser(description="Daily price comparison")
parser.add_argument("--test", "-t", action="store_true", help="Enable test mode")
args = parser.parse_args()

db_dir = "db"
if not os.path.exists(db_dir):
    os.makedirs(db_dir)

if args.test:
    db_path = os.path.join(db_dir, "test_products.db")
else:
    db_path = os.path.join(db_dir, "products.db")

conn = sqlite3.connect(db_path)

create_table_query = """
CREATE TABLE IF NOT EXISTS products (
    id TEXT,
    title TEXT,
    category TEXT,
    price TEXT,
    scrape_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    owned INTEGER,
    image TEXT
);
"""
conn.execute(create_table_query)
conn.commit()

df_db = pd.read_sql_query("SELECT * FROM products", conn)

if df_db.empty:
    json_file = "data/products.json"
    if not os.path.exists(json_file):
        print("No product data found in the database or JSON file. Exiting.")
        sys.exit(1)
    print("Database is empty. Loading data from JSON file...")
    try:
        with open(json_file, "r") as f:
            json_data = json.load(f)
    except Exception as e:
        print(f"Error reading JSON file: {e}")
        sys.exit(1)
    df_json = pd.DataFrame(json_data)
    if 'scrape_date' not in df_json.columns:
        current_time = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        df_json["scrape_date"] = current_time
    df_json.to_sql("products", conn, if_exists="append", index=False)
    print("Data loaded into the database from JSON.")
    df_db = pd.read_sql_query("SELECT * FROM products", conn)

unique_dates = df_db['scrape_date'].unique()
print("Unique scrape dates in database:", unique_dates)

if args.test and len(unique_dates) < 2:
    print("Test mode active: Simulating a second scrape date.")
    base_date = unique_dates[0]
    try:
        base_dt = datetime.datetime.strptime(base_date, "%Y-%m-%d %H:%M:%S")
    except Exception as e:
        print(f"Error parsing base date: {e}")
        sys.exit(1)
    new_date = (base_dt - datetime.timedelta(days=1)).strftime("%Y-%m-%d %H:%M:%S")
    df_simulated = df_db.copy()
    df_simulated['scrape_date'] = new_date
    df_simulated['price'] = df_simulated['price'].apply(lambda x: f"${float(x.replace('$','')) * random.uniform(0.9, 1.1):.2f}")
    df_simulated.to_sql("products", conn, if_exists="append", index=False)
    df_db = pd.read_sql_query("SELECT * FROM products", conn)
    unique_dates = df_db['scrape_date'].unique()
    print("After simulation, unique scrape dates:", unique_dates)

query_dates = "SELECT DISTINCT scrape_date FROM products ORDER BY scrape_date DESC LIMIT 2;"
dates = [row[0] for row in conn.execute(query_dates)]
if len(dates) < 2:
    print("Not enough data to perform a comparison.")
    sys.exit(1)

date1, date2 = dates[0], dates[1]
print(f"Comparing dates: {date1} (latest) and {date2} (previous)")

query_latest = f"SELECT * FROM products WHERE scrape_date = '{date1}'"
query_previous = f"SELECT * FROM products WHERE scrape_date = '{date2}'"

df_latest = pd.read_sql_query(query_latest, conn)
df_previous = pd.read_sql_query(query_previous, conn)

if 'id' not in df_latest.columns or 'id' not in df_previous.columns:
    print("id column missing in one of the datasets.")
    sys.exit(1)

comparison = pd.merge(df_latest, df_previous, on="id", suffixes=("_latest", "_previous"))

def clean_price(price):
    try:
        return float(str(price).replace("$", "").replace(",", ""))
    except Exception:
        return None

comparison["price_latest"] = comparison["price_latest"].apply(clean_price)
comparison["price_previous"] = comparison["price_previous"].apply(clean_price)
comparison["price_change"] = comparison["price_latest"] - comparison["price_previous"]

with pd.ExcelWriter("daily_price_comparison.xlsx") as writer:
    # Write all scraped items
    df_db.to_excel(writer, sheet_name="All Items", index=False)
    
    # Filter items with price changes (non-zero price_change)
    df_changes = comparison[comparison["price_change"] != 0]
    df_changes.to_excel(writer, sheet_name="Price Changes", index=False)

print("Comparison complete. Results saved to daily_price_comparison.xlsx")
conn.close()