-- 001_create_users.sql
-- Module: 01-auth (sdd/specs/modules/01-auth.md)
-- Applied automatically by the postgres container on first startup
-- (docker-entrypoint-initdb.d runs files in lexical order).

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(255)  NOT NULL,
    email         VARCHAR(255)  NOT NULL UNIQUE,
    password_hash VARCHAR(255)  NOT NULL,
    role          VARCHAR(10)   NOT NULL DEFAULT 'client' CHECK (role IN ('admin', 'client')),
    status        VARCHAR(10)   NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'banned', 'timeout')),
    timeout_until TIMESTAMPTZ,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
