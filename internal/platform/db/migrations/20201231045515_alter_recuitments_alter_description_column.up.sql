ALTER TABLE recruitments DROP COLUMN IF EXISTS description;
ALTER TABLE recruitments ADD COLUMN description text;