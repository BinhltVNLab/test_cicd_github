alter table users drop constraint if exists index_users_reset_password_code;
alter table users drop column reset_password_code;
alter table users drop column code_expired_at;
