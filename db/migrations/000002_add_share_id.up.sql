CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
ALTER TABLE list_shares ADD COLUMN id TEXT;

UPDATE list_shares SET id = uuid_generate_v4()::TEXT;

ALTER TABLE list_shares ALTER COLUMN id SET NOT NULL;
ALTER TABLE list_shares ALTER COLUMN id SET UNIQUE;
ALTER TABLE list_shares DROP CONSTRAINT list_shares_pkey;
ALTER TABLE list_shares ADD PRIMARY KEY (id);
