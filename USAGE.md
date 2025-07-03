## Service Design

The internal API runs on port :5665

`GET /events?product_id=...&limit=...`

This does need a basic auth header with the token 'dGVzdDp0ZXN0' (see below in retrieve events section)

Mock product ID's and their current quantities are as follows:
```
f47ac10b-58cc-4372-a567-0e02b2c3d479    11
8c138fa0-bfb4-4fd3-a23a-fed6468337d3    1
a4a00466-c889-4bec-9eb2-89fb4950da6c    0
b1c31e32-721b-472f-84bc-35a95d595184    5
119eefcc-0477-47d8-9f4b-6c55c030a78b    20
```

## Installation & Usage

**Install and set up services**

```
git clone https://github.com/jgarner9/inventory-events-service.git
cd inventory-events-service/inventory-events-service
sudo docker compose up --build
// this will start the RabbitMQ, PostgreSQL, and inventory-events-service instances
```

**Run the following to mock events to this service (use this to populate the DB to get responses on the API)**

```
cd inventory-events-service/mock-inventory-events
go run main.go <product_id> <mock_new_quantity>
// this will send a mock event to the inventory-events-service and update the quantities in-memory
```

**Run the following to retrieve events from the service**

```
curl -H "Authorization: Bearer dGVzdDp0ZXN0" http://localhost:5665?product_id=...&limit=...
// the bearer token is test:test base64 encoded to simulate internal authorization, as per requirements
```
