from flask import Flask, request, jsonify
import os
import shutil
import time  # Добавлено
app = Flask(__name__)

UPLOAD_DIR = "/Users/icetusk/Projects/ws-quiz/static/uploads"
PROCESSED_DIR = "/Users/icetusk/Projects/ws-quiz/static/processed"

os.makedirs(UPLOAD_DIR, exist_ok=True)
os.makedirs(PROCESSED_DIR, exist_ok=True)

@app.route('/process', methods=['POST'])
def process_files():
    print("🔔 Получен запрос:", request.method, request.headers)
    print("🔧 raw body:", request.data)
    print("🔧 parsed json:", request.get_json(force=True, silent=True))
    data = request.get_json()
    if not data or 'files' not in data or not isinstance(data['files'], list):
        return jsonify({"error": "Invalid input, expected JSON with 'files': [list]"}), 400

    failed = []
    for filename in data['files']:
        src_path = os.path.join(UPLOAD_DIR, filename)
        dst_path = os.path.join(PROCESSED_DIR, filename)

        if os.path.exists(src_path):
            try:
                shutil.copy(src_path, dst_path)
            except Exception as e:
                failed.append({"file": filename, "error": str(e)})
        else:
            failed.append({"file": filename, "error": "File not found"})

    time.sleep(10)  # ⏳ ЗАДЕРЖКА на 10 секунд для демонстрации !!!

    if failed:
        return jsonify({"status": "partial", "errors": failed}), 207

    return jsonify({"status": "ok"}), 200

if __name__ == '__main__':
    app.run(debug=True)