import sqlite3
import pandas as pd

# Connect to the SQLite database
conn = sqlite3.connect('db/products.db')

# Retrieve the two most recent scrape dates
query_dates = """
    SELECT DISTINCT scrape_date
    FROM products
    ORDER BY scrape_date DESC
    LIMIT 2;
"""
dates = [row[0] for row in conn.execute(query_dates)]
if len(dates) < 2:
    print("Not enough data to perform a comparison.")
    exit()

# Get data for the latest and previous dates
query_latest = f"SELECT * FROM products WHERE scrape_date = '{dates[0]}'"
query_previous = f"SELECT * FROM products WHERE scrape_date = '{dates[1]}'"

df_latest = pd.read_sql_query(query_latest, conn)
df_previous = pd.read_sql_query(query_previous, conn)

# Merge the two dataframes on SKU
comparison = pd.merge(
    df_latest, df_previous, 
    on='sku', 
    suffixes=('_latest', '_previous')
)

# Compute price differences
comparison['price_change'] = comparison['price_latest'] - comparison['price_previous']

# Separate increases and decreases if needed
price_increases = comparison[comparison['price_change'] > 0]
price_decreases = comparison[comparison['price_change'] < 0]

# Output the results to an Excel file
with pd.ExcelWriter('daily_price_comparison.xlsx') as writer:
    comparison.to_excel(writer, sheet_name='All Changes', index=False)
    price_increases.to_excel(writer, sheet_name='Price Increases', index=False)
    price_decreases.to_excel(writer, sheet_name='Price Decreases', index=False)

conn.close()