CREATE TABLE IF NOT EXISTS profiles(
   id public.xid PRIMARY KEY DEFAULT xid(),
   name VARCHAR (50) UNIQUE NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
   deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS checks(
    id public.xid PRIMARY KEY DEFAULT xid(),
    name VARCHAR (50) UNIQUE NOT NULL,
    upstream_services TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);


CREATE TABLE IF NOT EXISTS profiles_checks(
    profile_id public.xid REFERENCES profiles (id) ON UPDATE CASCADE ON DELETE CASCADE,
    check_id public.xid REFERENCES checks (id) ON UPDATE CASCADE,
    CONSTRAINT profiles_checks_pkey PRIMARY KEY (profile_id, check_id)
);
