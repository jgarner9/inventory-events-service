Project Based Learning Part 1:
Prompt given to ChatGPT = "pretend you are my boss at a software company (fill in the blanks on any details there yourself). give me the software requirements for a microservice you would like me to build for an existing architecture (again, make up any necessary details). i will use this to do more project based learning"

AI SLOP:
💼 Boss to Developer Brief
To: Jason Garner
From: Dana Hartley, VP of Engineering, NovaLoop Technologies
Subject: Software Requirements – Inventory Notifications Microservice
Date: June 19, 2025

📦 Project Background
NovaLoop Technologies is building out a modular, cloud-native e-commerce platform for mid-sized retail clients. Our backend is based on a microservice architecture running in a Kubernetes cluster, written primarily in Go and Node.js, with RabbitMQ for messaging, PostgreSQL for persistent storage, and Redis for caching. Services communicate over gRPC and REST, depending on use case.

Currently, our system lacks an efficient way to notify external systems and internal services when an inventory level crosses critical thresholds (e.g., low stock, out-of-stock, back in stock). We need a dedicated microservice to manage inventory threshold events and notify downstream systems such as the fulfillment engine, email dispatcher, and admin dashboard.

🧩 Your Assignment: inventory-events-service
Build a new microservice called inventory-events-service. This service is responsible for:

Subscribing to inventory updates from the inventory-service.
Evaluating threshold conditions.
Publishing events to a RabbitMQ topic for other services to consume.
Logging events and their outcomes to a PostgreSQL table.
Providing an internal API (REST or gRPC) to query recent inventory events for auditing purposes.
🛠️ Technical Requirements
Language & Framework

Primary language: Go (Golang)
Frameworks: Use Gin or chi for REST (if applicable), or grpc-go for gRPC
Database ORM: gorm or raw SQL with pgx
Messaging

Broker: RabbitMQ (via AMQP 0.9.1)
Input queue: inventory.updates
Output topic: inventory.events
Database

Type: PostgreSQL (via internal postgres-service)
Schema:
CREATE TABLE inventory_events (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    previous_quantity INTEGER,
    new_quantity INTEGER,
    event_type TEXT, -- 'LOW_STOCK', 'OUT_OF_STOCK', 'BACK_IN_STOCK'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
Threshold Logic

You will apply the following business rules:

LOW_STOCK: when quantity drops below 10 but is above 0
OUT_OF_STOCK: when quantity reaches 0
BACK_IN_STOCK: when quantity goes from 0 to any positive number
This logic must only trigger once per transition. Don't repeatedly fire LOW_STOCK on every message unless it’s a fresh state change.
Internal API

Endpoint: GET /events?product_id=<uuid>&limit=20
Returns: Most recent inventory_events for a given product ID in JSON format
Auth: Internal token-based authentication (we'll simulate this with a simple header check: Authorization: Bearer internal-token)
🧪 Testing & Local Setup
Use Docker for local RabbitMQ and PostgreSQL instances
Provide a docker-compose.yml with all dependencies
Write a few integration tests to simulate inventory changes and assert the correct events fire
📦 Deliverables
inventory-events-service source code in Go
Dockerfile + docker-compose.yml
README.md with setup instructions
Example curl or grpcurl commands for testing
Integration tests
Let me know if you need additional help around RabbitMQ setup, event idempotency, or if you want to pair program some of the logic. Otherwise, looking forward to your initial prototype!
