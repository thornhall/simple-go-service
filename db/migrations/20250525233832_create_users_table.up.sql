CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
  id          BIGSERIAL PRIMARY KEY,           
  object_id   UUID NOT NULL DEFAULT uuid_generate_v4(), 
  is_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
  first_name  TEXT    NOT NULL,
  last_name   TEXT,
  email       TEXT    NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT users_object_id_key UNIQUE(object_id)
);
