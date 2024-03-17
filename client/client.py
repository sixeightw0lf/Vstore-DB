import json
from flask import Flask, render_template
from flask_socketio import SocketIO, emit, join_room, leave_room, close_room, rooms, disconnect
import requests
from urllib.parse import quote


app = Flask(__name__)
app.config['SECRET_KEY'] = 'your-secret-key'
socketio = SocketIO(app)

API_URL = 'http://localhost:8080'

@app.route('/')
def index():
    return render_template('index.html')

@socketio.on('connect')
def handle_connect():
    print('Client connected')

@socketio.on('disconnect')
def handle_disconnect():
    print('Client disconnected')

@socketio.on('get_all_records')
def handle_get_all_records():
    print("get_all_records")
    try:
        response = requests.get(f'{API_URL}/get/all')
        data = response.json()
        emit('records_updated', data, broadcast=True)
    except Exception as e:
        emit('error', {'error': str(e)}, broadcast=True)

@socketio.on('get_record_by_id')
def handle_get_record_by_id(id):
    response = requests.get(f'{API_URL}/get/{id}')
    if response.status_code == 200:
        try:
            data = response.json()
            emit('record_updated', data)
        except json.JSONDecodeError:
            emit('error', 'Invalid JSON response from the server')
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('get_record_by_id_and_pivot')
def handle_get_record_by_id_and_pivot(id, pivot_key):
    response = requests.get(f'{API_URL}/get/{id}/{pivot_key}')
    if response.status_code == 200:
        try:
            data = response.json()
            emit('record_updated', data)
        except json.JSONDecodeError:
            emit('error', 'Invalid JSON response from the server')
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('search_records')
def handle_search_records(keyword, fuzzy):
    url = f'{API_URL}/search/{keyword}'
    if fuzzy:
        url += '/fuzzy'
    response = requests.get(url)
    if response.status_code == 200:
        try:
            data = response.json()
            emit('records_updated', data)
        except json.JSONDecodeError:
            emit('error', 'Invalid JSON response from the server')
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('query_records')
def handle_query_records(keywords, fuzzy):
    url = f'{API_URL}/query/{"/".join(keywords)}'
    if fuzzy:
        url += '?fuzzy=true'
    response = requests.get(url)
    if response.status_code == 200:
        try:
            data = response.json()
            emit('records_updated', data)
        except json.JSONDecodeError:
            emit('error', 'Invalid JSON response from the server')
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('add_record')
def handle_add_record(record):
    response = requests.post(f'{API_URL}/data', json=record)
    if response.status_code == 200:
        emit('record_added', record, broadcast=True)
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('update_record')
def handle_update_record(record):
    response = requests.post(f'{API_URL}/data', json=record)
    if response.status_code == 200:
        emit('record_updated', record, broadcast=True)
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('delete_record')
def handle_delete_record(record_id):
    response = requests.delete(f'{API_URL}/data/{record_id}')
    if response.status_code == 200:
        emit('record_deleted', record_id, broadcast=True)
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('set_encryption_key')
def handle_set_encryption_key(password):
    response = requests.post(f'{API_URL}/password', json={'password': password})
    if response.status_code == 200:
        emit('encryption_key_set')
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('connect_database')
def handle_connect_database():
    response = requests.post(f'{API_URL}/connect')
    if response.status_code == 200:
        emit('database_connected')
    else:
        emit('error', f'Request failed with status code {response.status_code}')

@socketio.on('disconnect_database')
def handle_disconnect_database():
    response = requests.post(f'{API_URL}/disconnect')
    if response.status_code == 200:
        emit('database_disconnected')
    else:
        emit('error', f'Request failed with status code {response.status_code}')

if __name__ == '__main__':
    socketio.run(app, port=5111)