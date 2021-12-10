alter table users add column reset_password_code text;
alter table users add column code_expired_at timestamp;
create index index_users_reset_password_code on users (reset_password_code);
