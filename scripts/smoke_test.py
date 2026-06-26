#!/usr/bin/env python3
import json
import os
import sys
import time
import urllib.error
import urllib.request


BASE_URL = os.environ.get("BASE_URL", "http://localhost:8000").rstrip("/")
TIMEOUT_SECONDS = float(os.environ.get("TIMEOUT_SECONDS", "5"))
RETRIES = int(os.environ.get("RETRIES", "12"))
SLEEP_SECONDS = float(os.environ.get("SLEEP_SECONDS", "2"))


def request(path: str, method: str = "GET", payload: dict | None = None) -> tuple[int, str]:
    data = None
    headers = {}

    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(f"{BASE_URL}{path}", data=data, headers=headers, method=method)

    with urllib.request.urlopen(req, timeout=TIMEOUT_SECONDS) as response:
        return response.status, response.read().decode("utf-8", errors="replace")


def wait_for_api() -> None:
    last_error = None

    for attempt in range(1, RETRIES + 1):
        try:
            status, _ = request("/api/v1/healthcheck/")
            if 200 <= status < 300:
                print(f"OK: healthcheck returned {status}")
                return
        except (urllib.error.URLError, TimeoutError) as exc:
            last_error = exc

        print(f"Waiting for API ({attempt}/{RETRIES})...")
        time.sleep(SLEEP_SECONDS)

    raise RuntimeError(f"API did not become healthy: {last_error}")


def main() -> int:
    wait_for_api()

    status, metrics = request("/api/v1/metrics")
    if status != 200 or "http_requests_total" not in metrics:
        raise RuntimeError("metrics endpoint is not exposing expected Prometheus metrics")
    print("OK: metrics endpoint exposes Prometheus metrics")

    payload = {
        "name": "Smoke Test Car",
        "colour": "#1122AA",
        "price": 12345,
        "build_date": "2024-05-01",
    }
    status, body = request("/api/v1/cars/c/", method="POST", payload=payload)
    if status not in (200, 201, 202):
        raise RuntimeError(f"create car returned unexpected status {status}: {body}")
    print(f"OK: create car returned {status}")

    status, body = request("/api/v1/cars/q/")
    if status != 200:
        raise RuntimeError(f"list cars returned unexpected status {status}: {body}")
    print("OK: list cars returned 200")

    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        raise SystemExit(1)
