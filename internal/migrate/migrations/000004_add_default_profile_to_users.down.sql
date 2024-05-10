-- Drop the foreign key constraint
ALTER TABLE users
    DROP CONSTRAINT IF EXISTS fk_default_profile;

-- Drop the default_profile column
ALTER TABLE users
    DROP COLUMN IF EXISTS default_profile;
