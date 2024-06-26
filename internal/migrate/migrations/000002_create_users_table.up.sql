CREATE TABLE IF NOT EXISTS users(
   id public.xid PRIMARY KEY DEFAULT xid(),
   name VARCHAR (50) UNIQUE NOT NULL,
   email VARCHAR (300) UNIQUE NOT NULL,
   token BYTEA NOT NULL,
   is_admin BOOLEAN NOT NULL DEFAULT FALSE,
   created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
   deleted_at TIMESTAMPTZ
);
