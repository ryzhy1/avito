-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE bids_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bid_id UUID NOT NULL,
    version INTEGER NOT NULL,
    name VARCHAR(255),
    status VARCHAR(20),
    author_type VARCHAR(100),
    author_id INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_bid
        FOREIGN KEY (bid_id)
        REFERENCES bids(id)
);

-- +goose Down
DROP TABLE bids_versions;
