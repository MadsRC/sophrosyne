-- Step 1: Add the default_profile column
ALTER TABLE users
    ADD COLUMN default_profile public.xid;

-- Step 2: Add a foreign key constraint to the default_profile column
ALTER TABLE users
    ADD CONSTRAINT fk_default_profile FOREIGN KEY (default_profile) REFERENCES profiles (id);
