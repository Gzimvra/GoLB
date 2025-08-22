# Custom Load Balancer in Go

## Project Overview
This is a custom TCP/HTTP load balancer written in Go.
It accepts client requests, distributes them across multiple backend servers using simple algorithms (e.g., round-robin), and keeps track of server health with background checks.

The project is designed as a learning exercise to explore how load balancers work under the hood, including connection handling, request forwarding, and health monitoring.

---

# Architecture & Request Flow

This document explains the request flow through the load balancer, what gets returned to the user, and how health checks are managed.

---

## Request Lifecycle

### 1. Client Sends a Request
- Example: User hits `https://myapp.com/api/users`.
- DNS resolves `myapp.com` → IP address of your load balancer.

---

### 2. Load Balancer Receives the Request
- Accepts the TCP/HTTP connection.
- (Optional) **TLS termination**: LB decrypts HTTPS and forwards plain HTTP to backends.
- (Optional) **Rate limiting / DDoS protection**: LB may throttle abusive clients.
- Reads the request (headers, body, etc.).

---

### 3. Load Balancing Decision
- **Health checks**: The LB maintains a pool of healthy servers and ignores unhealthy ones.
- **Algorithm**: Selects a backend (e.g., round robin, least connections, random).
- (Optional: Depends on the algorithm used) **Session persistence (sticky sessions)**: Same user is routed to the same backend.
- Example: LB chooses `http://app-server-2:9000`.

---

### 4. Forwarding the Request (Reverse Proxying)
- The LB forwards the request to the chosen backend:
  - **HTTP**: Adjusts headers (`X-Forwarded-For`, `Host`) so the backend knows the real client.
  - **TCP**: Simply forwards raw data.
- (Optional) **Retries / failover**: If a backend fails, the LB retries with another healthy server.

---

### 5. Backend Application Server Handles the Request
- Processes the request (e.g., database queries, business logic).
- Returns an HTTP response (status code, headers, body).

---

### 6. Load Balancer Receives the Response
- Gets the response from the backend.
- (Optional) Modifies or adds headers (e.g., `X-Backend-Server: app-server-2` for debugging).
- (Optional) Applies optimizations such as compression or caching.

---

### 7. Client Receives the Response
- To the user, it looks like the response came directly from the load balancer (they don’t see which backend handled it).
- Example:
  ```json
  [{ "id": 1, "name": "Mark" }]
  ```

---

## What Gets Returned to the User?
- The backend’s response (status + headers + body).
- Possibly extra headers added by the load balancer (for logging/tracing).
- If the backend is down and no healthy servers are available:
    - LB can return an error (`502 Bad Gateway`, `503 Service Unavailable`).

## What About Health Checks — When & Where
### 1. When Do Health Checks Run?
- They don’t happen on every client request (that would add latency and overhead).
- Instead, the load balancer runs them in the background on a schedule (e.g., every 5 seconds, 10 seconds, etc.) via a separate goroutine.
- Health checks are proactive: the LB keeps an up-to-date view of which servers are alive before traffic arrives.

### 2. Where Do Health Checks Happen?
- A separate goroutine inside the LB periodically sends test requests to each backend.
    - HTTP: The LB calls `http://server:port/health` or `/ping` and expects `200 OK`.
    - TCP: The LB tries opening a socket connection. If the handshake succeeds, the server is alive.

### 3. How Are Results Used?
- The LB keeps a pool of backend servers with status info like: "Alive"
- The health checker updates `Alive = true/false` in that pool.
- When a client request comes in (step 3 of the flow), the load balancing algorithm only considers servers where `Alive == true`.

