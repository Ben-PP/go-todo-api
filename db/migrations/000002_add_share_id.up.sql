BEGIN;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
ALTER TABLE list_shares ADD COLUMN id TEXT;

DO $$
DECLARE
    share RECORD;
BEGIN
    FOR share IN SELECT * FROM list_shares LOOP
        UPDATE list_shares SET id = uuid_generate_v4() WHERE list_id = share.list_id AND user_id = share.user_id;
    END LOOP;
END $$;

ALTER TABLE list_shares DROP CONSTRAINT list_shares_pkey;
ALTER TABLE list_shares ADD PRIMARY KEY (id);
COMMIT;