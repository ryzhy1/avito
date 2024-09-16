-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    service_type VARCHAR(255),
    organization_id UUID NOT NULL,
    creator_username VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT fk_organization
        FOREIGN KEY (organization_id)
        REFERENCES organization(id)
);

-- +goose Down
DROP TABLE tenders;