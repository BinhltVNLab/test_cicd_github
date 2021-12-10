ALTER TABLE user_permissions DROP COLUMN IF EXISTS modules;
ALTER TABLE user_permissions ADD COLUMN function_id int;
ALTER TABLE user_permissions ADD COLUMN status smallint not null default 2;
alter table user_permissions add constraint user_permission_function_id foreign key (function_id) references functions (id);
