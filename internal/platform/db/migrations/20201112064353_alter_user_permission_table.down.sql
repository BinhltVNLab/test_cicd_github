ALTER TABLE user_permissions DROP COLUMN IF EXISTS function_id;
ALTER TABLE user_permissions DROP COLUMN IF EXISTS status;
alter table user_permissions drop constraint if exists user_permission_function_id;
