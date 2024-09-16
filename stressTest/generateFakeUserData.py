import os
import sys
import mysql.connector
from faker import Faker
from dotenv import load_dotenv
from tqdm import tqdm
import csv
import base64
import hashlib

# Load environment variables from .env file
load_dotenv()

# Retrieve database configuration from environment variables
config = {
    'user': os.getenv('DB_USER'),
    'password': os.getenv('DB_PASSWORD'),
    'host': os.getenv('DB_HOST'),
    'database': os.getenv('DB_NAME')
}

# Create a Faker instance
fake = Faker()
existing_usernames = set()


def generate_unique_user_name():
    while True:
        user_name = fake.user_name()
        if user_name not in existing_usernames:
            existing_usernames.add(user_name)
            return user_name


def generate_user_data(number):
    """Generates fake user data."""
    password = fake.password(length=12)
    salt = fake.word()
    return (
        hash_password(password, salt),  # hashed_password
        salt,  # salt
        fake.first_name(),  # first_name
        fake.last_name(),  # last_name
        fake.date_of_birth(minimum_age=18, maximum_age=90).today(),  # dob
        fake.email(),  # email
        base64_encode_number(number),
        password
    )


def hash_password(password: str, salt: str) -> str:
    # Combine the password and salt
    password_salt = password + salt

    # Create a SHA-256 hash object
    hash_object = hashlib.sha256()

    # Update the hash object with the combined password and salt
    hash_object.update(password_salt.encode('utf-8'))

    # Get the binary hash value
    hash_bytes = hash_object.digest()

    # Encode the hash value in Base64
    hash_base64 = base64.b64encode(hash_bytes).decode('utf-8')

    return hash_base64


def insert_users(lowerBound, upperBound, batch_size=100000):
    """Inserts `num_users` into the user table."""
    conn = mysql.connector.connect(**config)
    cursor = conn.cursor()

    insert_query = """
    INSERT INTO user (hashed_password, salt, first_name, last_name, dob, email, user_name)
    VALUES (%s, %s, %s, %s, %s, %s, %s)
    """

    with open('users.csv', 'w', newline='') as csvfile:
        user_writer = csv.writer(csvfile)
        user_writer.writerow(['username', 'password'])
        batch = []
        for (num, progress) in zip(range(lowerBound, upperBound + 1, 1), tqdm(range(lowerBound, upperBound + 1, 1))):
            user_data = generate_user_data(num)
            batch.append(user_data[:-1])  # Exclude password for DB insertion
            user_writer.writerow(user_data[6:8])  # Write username and password to CSV
            if len(batch) >= batch_size:
                cursor.executemany(insert_query, batch)
                conn.commit()
                batch = []
        if batch:
            cursor.executemany(insert_query, batch)
            conn.commit()
    cursor.close()
    conn.close()


def base64_encode_number(number: int) -> str:
    return str(number)

if __name__ == "__main__":
    lowerBound = int(sys.argv[1])
    upperBound = int(sys.argv[2])
    insert_users(lowerBound=lowerBound, upperBound=upperBound)  # Insert 10 million users
