import random
import mysql.connector
from tqdm import tqdm
from dotenv import load_dotenv
import os
import sys

# Load environment variables from .env file
load_dotenv()

# Define constants
TOTAL_USERS = 10_000_000
USER_LOWER_BOUND_ID = 11_878_102  # Starting user_id
USER_UPPER_BOUND_ID = 21_878_101  # Ending user_id
AVG_FOLLOWERS = 200
SPECIAL_USERS_COUNT = 1_000  # Number of special users
SPECIAL_USER_FOLLOWERS = 100_000  # Number of followers for special users
BATCH_SIZE = 10

# Retrieve database configuration from environment variables
config = {
    'user': os.getenv('DB_USER'),
    'password': os.getenv('DB_PASSWORD'),
    'host': os.getenv('DB_HOST'),
    'database': os.getenv('DB_NAME')
}
conn = mysql.connector.connect(**config)
cursor = conn.cursor()

# Function to generate random follower user_ids
def generate_followers(user_id, follower_count):
    followers = set()

    while len(followers) < follower_count:
        follower_id = random.randint(USER_LOWER_BOUND_ID, USER_UPPER_BOUND_ID)
        if follower_id != user_id:
            followers.add(follower_id)

    return followers

# Insert follower records into the database
def insert_followers(user_id, followers):
    follower_ids = [(user_id, follower) for follower in followers]

    if follower_ids:
        sql = "INSERT INTO user_user (fk_user_id, fk_follower_id) VALUES (%s, %s)"
        cursor.executemany(sql, follower_ids)
        conn.commit()
        print(f"Successfully inserted {len(follower_ids)} followers for user id {user_id}")

def insert_follower(userIDStart):
    # Generate followers for all users
    try:
        cnt = 0
        for (user_id, progress) in zip(range(userIDStart, userIDStart + 1_000_000), tqdm(range(0, 1_000_000))):
            cnt += 1
            # Check if the user_id already exists in the user_user table
            cursor.execute("SELECT COUNT(*) FROM user_user WHERE fk_user_id = %s", (user_id,))
            result = cursor.fetchone()

            # Skip user if they already have followers
            if result[0] > 0:
                print(f"User {user_id} already has followers, skipping.")
                continue
            # For special users, generate a large number of followers
            if cnt % 1000 == 0:
                followers_count = SPECIAL_USER_FOLLOWERS
            else:
                followers_count = AVG_FOLLOWERS

            # Generate followers
            new_followers = generate_followers(user_id, followers_count)

            # Insert new followers into the database
            insert_followers(user_id, new_followers)
        # Final commit for any remaining inserts
        conn.commit()
        print("Follower generation completed.")
    except Exception as e:
        print(f"Error: {e}")

    finally:
        cursor.close()
        conn.close()

if __name__ == "__main__":
    if len(sys.argv) > 1:
        # Use eval() to evaluate the expression passed as a string
        userIDStart = eval(sys.argv[1])
        print(f"User ID Start: {userIDStart}")
        insert_follower(userIDStart=userIDStart)
    else:
        print("Please provide an expression as an argument.")
