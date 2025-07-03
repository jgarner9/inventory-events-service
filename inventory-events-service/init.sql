CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS inventory_events (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id UUID NOT NULL,
	previous_quantity INTEGER,
	new_quantity INTEGER,
	event_type TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
