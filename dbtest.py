import requests

# Base URL of the Go-based database API
BASE_URL = "http://localhost:8080"

def save_row(key, value):
    """Saves a single row to the database."""
    response = requests.post(BASE_URL, json={'key': key, 'value': value})
    print(f"Save {key}: {response.text}")

def get_row(key):
    """Retrieves a single row from the database by key."""
    response = requests.get(f"{BASE_URL}?key={key}")
    return response.text

def main():
    row = get_row('1')
    print(row)
    # # Example rows to save
    # rows_to_save = [
    #     ('1', 'test'),
    #     ('2', 'hello'),
    #     ('3', 'world'),
    # ]

    # # Save rows to the database
    # for key, value in rows_to_save:
    #     save_row(key, value)

    # # Retrieve and print all saved rows
    # for key, _ in rows_to_save:
    #     value = get_row(key)
    #     print(f"Get {key}: {value}")

if __name__ == "__main__":
    main()
