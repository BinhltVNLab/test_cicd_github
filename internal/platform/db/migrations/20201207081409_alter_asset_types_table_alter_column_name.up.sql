ALTER TABLE asset_types DROP COLUMN IF EXISTS name;
ALTER TABLE asset_types ADD COLUMN name varchar (150) unique;