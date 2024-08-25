import mysql.connector
import csv
from dotenv import load_dotenv
import os

# Load environment variables from .env file
load_dotenv()

# Retrieve database configuration from environment variables
config = {
    'user': os.getenv('DB_USER'),
    'password': os.getenv('DB_PASSWORD'),
    'host': os.getenv('DB_HOST'),
    'database': os.getenv('DB_NAME')
}

# Number of rows to fetch
NUM_ROWS = 100_000
CHUNK_SIZE = 10_000  # Number of rows to fetch in each chunk

# Connect to the database
conn = mysql.connector.connect(**config)
cursor = conn.cursor()

# Prepare the query to fetch usernames and passwords
query = "SELECT user_name, hashed_password FROM user LIMIT %s OFFSET %s"

# Open a file to write user data
with open('users.csv', 'w', newline='') as csvfile:
    user_writer = csv.writer(csvfile)
    user_writer.writerow(['username', 'password'])  # Write headers

    offset = 0
    rows_written = 0

    while rows_written < NUM_ROWS:
        cursor.execute(query, (CHUNK_SIZE, offset))
        rows = cursor.fetchall()

        # Break the loop if no more rows are returned
        if not rows:
            break

        # Write the fetched data to CSV
        user_writer.writerows(rows)
        rows_written += len(rows)
        offset += CHUNK_SIZE

# Close the cursor and connection
cursor.close()
conn.close()