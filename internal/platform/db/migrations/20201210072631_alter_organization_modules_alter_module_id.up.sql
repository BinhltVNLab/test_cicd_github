alter table functions drop constraint if exists function_module_id;
ALTER TABLE organization_modules DROP COLUMN IF EXISTS module_id;
ALTER TABLE organization_modules ADD COLUMN modules jsonb;