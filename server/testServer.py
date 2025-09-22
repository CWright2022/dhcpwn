#!/usr/bin/env python3
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route("/checkin", methods=["POST"])
def checkin():
    if not request.is_json:
        return jsonify({"error": "expected application/json"}), 400

    req = request.get_json()
    if req is None:
        return jsonify({"error": "invalid or missing JSON body"}), 400
    client_id = req.get("clientID", "test-client")

    # Static payload that matches Go Payload struct
    response_payload = {
        "clientID": client_id,     # echo clientID if provided
        "command": "run",          # static action
        "commandID": "0",          # string, unique ID in real server
        "args": "echo hello",      # arguments
    }

    return jsonify(response_payload), 200


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8000, debug=True)
