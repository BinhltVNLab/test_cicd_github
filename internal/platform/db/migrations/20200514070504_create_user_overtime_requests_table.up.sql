create table if not exists user_overtime_requests(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    user_id integer not null,
    project_id integer not null,
    status integer not null,
    datetime_overtime_from timestamp not null,
    datetime_overtime_to timestamp not null,
    email_title varchar(100),
    email_content varchar(1024),
    reason varchar(50),
    overtime_type integer,
    send_to varchar [],
    send_cc varchar []
);

create index index_ot_user_id on user_overtime_requests (user_id);
create index index_ot_project_id on user_overtime_requests (project_id);
create index index_datetime_overtime_from on user_overtime_requests (datetime_overtime_from);
create index index_datetime_overtime_to on user_overtime_requests (datetime_overtime_to);
create index index_datetime_overtime_from_to on user_overtime_requests (datetime_overtime_from, datetime_overtime_to);

alter table user_overtime_requests add constraint ot_user_id foreign key (user_id) references users (id);
alter table user_overtime_requests add constraint ot_project_id foreign key (project_id) references projects (id);

comment on column user_overtime_requests.id is 'bonus_leave_type id';
comment on column user_overtime_requests.created_at is 'Save timestamp when create';
comment on column user_overtime_requests.updated_at is 'Save timestamp when update';
comment on column user_overtime_requests.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_overtime_requests.user_id is 'Id of user';
comment on column user_overtime_requests.status is 'Overtime status 1: Pending, 2: Deny, 3: Accept';
comment on column user_overtime_requests.datetime_overtime_from is 'Datetime overtime from register';
comment on column user_overtime_requests.datetime_overtime_to is 'Datetime overtime to register';
comment on column user_overtime_requests.email_title is 'Title of email';
comment on column user_overtime_requests.email_content is 'Content of email';
comment on column user_overtime_requests.reason is 'Reason overtime';
comment on column user_overtime_requests.overtime_type is 'Type overtime';
comment on column user_overtime_requests.send_to is 'Send to';
comment on column user_overtime_requests.send_cc is 'Cc';
