ALTER TABLE list_shares DROP CONSTRAINT list_shares_pkey;
ALTER TABLE list_shares ADD PRIMARY KEY (list_id, user_id);
ALTER TABLE list_shares DROP COLUMN id;
DROP EXTENSION IF EXISTS "uuid-ossp";