alter table users drop constraint if exists index_users_update_email_code;
alter table users drop column email_for_update;
alter table users drop column update_email_code;
alter table users drop column update_email_code_expired_at;
