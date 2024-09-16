-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    author_type VARCHAR(100),
    author_id INTEGER,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    tender_id UUID NOT NULL,
    CONSTRAINT fk_tender
        FOREIGN KEY (tender_id)
        REFERENCES tenders(id)
);

-- +goose Down
DROP TABLE bids;
