alter table users add column email_for_update varchar(100);
alter table users add column update_email_code text;
alter table users add column update_email_code_expired_at timestamp;
create index index_users_update_email_code on users (update_email_code);
