# 📦 Inventory Notifications Microservice

## Project Based Learning

**Prompt to ChatGPT**
> _"Pretend you are my boss at a software company (fill in the blanks on any details there yourself). Give me the software requirements for a microservice you would like me to build for an existing architecture (again, make up any necessary details). I will use this to do more project-based learning."_

---

## 💼 Boss to Developer Brief

**To:** Jason Garner
**From:** Dana Hartley, VP of Engineering, *NovaLoop Technologies*
**Subject:** Software Requirements – Inventory Notifications Microservice
**Date:** June 19, 2025

---

## 📦 Project Background

NovaLoop Technologies is building a **modular, cloud-native e-commerce platform** for mid-sized retail clients. Our backend architecture includes:

- Microservices deployed in **Kubernetes**
- Services in **Go** and **Node.js**
- **RabbitMQ** for messaging
- **PostgreSQL** for persistent storage
- **Redis** for caching
- Inter-service communication using **gRPC** and **REST**

Currently, the system lacks a centralized way to handle inventory threshold events. We need a new microservice to monitor inventory levels and notify internal and external systems when critical thresholds are crossed.

---

## 🧩 Your Assignment: `inventory-events-service`

You are to build a microservice that:

- Subscribes to inventory updates from the `inventory-service`
- Detects when inventory crosses critical thresholds:
  - **Low Stock**
  - **Out of Stock**
  - **Back In Stock**
- Publishes events to a RabbitMQ topic
- Logs these events in PostgreSQL
- Provides an internal API for event auditing

---

## 🛠️ Technical Requirements

### 🔤 Language & Frameworks

- **Language:** Go (Golang)
- **REST (if applicable):** [Gin](https://github.com/gin-gonic/gin) or [chi](https://github.com/go-chi/chi)
- **gRPC (if applicable):** [grpc-go](https://github.com/grpc/grpc-go)
- **ORM:** [gorm](https://gorm.io/) or `pgx` with raw SQL

### 📨 Messaging (RabbitMQ)

- **Broker:** RabbitMQ (AMQP 0.9.1)
- **Input Queue:** `inventory.updates`
- **Output Topic:** `inventory.events`

### 🗃️ Database (PostgreSQL)

Use the internal `postgres-service`. Suggested schema:

```sql
CREATE TABLE inventory_events (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    previous_quantity INTEGER,
    new_quantity INTEGER,
    event_type TEXT, -- 'LOW_STOCK', 'OUT_OF_STOCK', 'BACK_IN_STOCK'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

### ⚙️ Threshold Logic

Apply the following rules:

**LOW_STOCK**: Quantity drops below 10 but above 0
**OUT_OF_STOCK**: Quantity reaches 0
**BACK_IN_STOCK**: Quantity increases from 0 to any positive number
> 🔁 Only trigger each event once per state transition — do not emit LOW_STOCK repeatedly if quantity stays below 10.

---

### 📡 Internal API

**Endpoint:**
`GET /events?product_id=<uuid>&limit=20`
**Returns:**
JSON list of recent inventory_events for a given product ID
Authentication:
Simulated internal token-based auth via header:
Authorization: Bearer internal-token

---

### 🧪 Testing & Local Setup

Use **Docker** to run local instances of RabbitMQ and PostgreSQL
Provide a `docker-compose.yml` to spin up dependencies
Write integration tests that:
Simulate inventory updates
Validate correct events are published and stored

---

### 📦 Deliverables

✅ `inventory-events-service` source code (written in Go)
✅ `Dockerfile` + `docker-compose.yml`
✅ `README.md` with setup instructions
✅ Example `curl` or `grpcurl` commands
✅ Integration tests

---

## 🤝 Notes

Let me know if you need assistance with:

Setting up RabbitMQ queues and topics
Ensuring event idempotency
Pair programming any part of the business logic
