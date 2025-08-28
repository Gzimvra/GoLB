# GoLB – Raw TCP Load Balancer in Go

## Project Overview

This project is a custom TCP load balancer written in Go.
It accepts incoming TCP connections from clients and distributes them across multiple backend servers using a simple algorithm (round-robin).

The load balancer also includes:

-   Health checks: backends are marked alive/unhealthy automatically.
-   IP whitelisting: restrict which clients can connect.
-   Rate limiting: protect against abusive clients.
-   Logging: connections, rejections, and errors are logged.

This project is mainly a learning exercise to explore how Layer 4 load balancers work under the hood.

## How to run

### 1. Clone and Build

```bash
git clone https://github.com/Gzimvra/golb.git
cd golb
go build -o golb ./cmd/golb
```

### 2. Start Example Backends

Open 3 separate terminals and run the provided backend servers:

```bash
go run ./examples/backends/backend1/main.go
go run ./examples/backends/backend2/main.go
go run ./examples/backends/backend3/main.go
```

By default, they listen on different ports (configured inside their main.go).

### 3. Configure the Load Balancer

-   Edit config.json if needed.

```json
{
    "listen_addr": ":8080",
    "algorithm": "round_robin",
    "health_check_interval": 10,
    "request_timeout": 5,
    "max_concurrent_connections": 10,
    "max_connections_per_minute": 50,
    "ip_filter_mode": "allow",
    "ip_filter_list": ["127.0.0.1"],
    "servers": [
        { "address": "localhost:9001" },
        { "address": "localhost:9002" },
        { "address": "localhost:9003" }
    ]
}
```

### 4. Run the Load Balancer

```bash
# run the binary
./golb

# or you can also run directly for convenience
go run ./cmd/golb
```

It will listen on :8080 (default) and forward connections to healthy backends.

### 5. Test from Client

From another terminal, use curl or nc:

```bash
curl http://127.0.0.1:8080
# or
nc 127.0.0.1 8080
```

Requests will be distributed to your backends in round-robin order.

## Architecture & Connection Lifecycle

This document explains the request flow through the load balancer, what gets returned to the user, and how health checks are managed.

### 1. Client Opens a TCP Connection

-   Example: User connects to `127.0.0.1:8080`.
-   The load balancer listens on a given address/port (config.json).

### 2. Load Balancer Receives the Request

-   Accepts the TCP connection.
-   (Configurable) IP filtering: connection is rejected if client IP is not allowed.
-   (Configurable) Rate limiting: connection is rejected if limits are exceeded.

### 3. Backend Selection

-   The load balancer maintains a pool of healthy servers.
-   It selects a backend using round-robin (default algorithm).
-   Only servers marked Alive == true are considered.

### 4. Connection Proxying

-   The LB proxies the TCP stream between the client and backend.
-   No application-level parsing or modification is performed.
-   Data is forwarded transparently in both directions.

### 5. Backend Responds

-   The backend processes the request and sends data back over the TCP connection.
-   The load balancer relays this data to the client.
-   To the user, it looks like the response came directly from the load balancer (they don’t see which backend handled it).
-   Example:
    ```json
    [{ "id": 1, "name": "Mark" }]
    ```

### 6. Connection Lifecycle Ends

-   When the client or backend closes the connection, the LB cleans up resources.
-   Metrics (total connections, rejections, active connections) can be tracked.

## What Gets Returned to the User?

-   Whatever the backend server sends.
-   The LB does not modify the payload — it simply forwards raw TCP data.
-   If your backends are HTTP servers, the client will see full HTTP responses.
-   For non-HTTP TCP apps, the raw stream is passed through unchanged.

## What About Health Checks — When & Where

### 1. When Do Health Checks Run?

-   They don’t happen on every client request (that would add latency and overhead).
-   Instead, the load balancer runs them in the background on a schedule (e.g., every 5 seconds, 10 seconds, etc.) via a separate goroutine.
-   Health checks are proactive: the LB keeps an up-to-date view of which servers are alive before traffic arrives.

### 2. Where Do Health Checks Happen?

-   A separate goroutine inside the LB periodically sends test requests to each backend.
-   The LB tries opening a socket connection. If the handshake succeeds, the server is alive.

### 3. How Are Results Used?

-   The LB keeps a pool of backend servers with status info like: "Alive"
-   The health checker updates `Alive = true/false` in that pool.
-   When a client request comes in (step 3 of the flow), the load balancing algorithm only considers servers where `Alive == true`.

## Future Enhancements

Some features that could be added to make the load balancer more production-ready:

-   TLS support for encrypted connections (client ↔ LB).
-   Sticky sessions to keep a client pinned to the same backend.
-   Pluggable algorithms (least-connections, random, weighted round-robin).
-   Prometheus metrics for observability.
-   Configuration reloads without downtime (hot reload).
