-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
);

CREATE TABLE organization (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type organization_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE organization;
DROP TABLE organization_responsible