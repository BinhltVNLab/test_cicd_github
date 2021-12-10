create table if not exists user_leave_requests(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,

    organization_id int not null,
    user_id int not null,
    leave_request_type_id int not null,
    datetime_leave_from timestamp not null,
    datetime_leave_to timestamp,
    created_by int not null,
    updated_by int not null,
    email_title varchar(100) not null,
    email_content varchar(255) not null,
    reason varchar(255) not null,
    hour real
);

comment on column user_leave_requests.id is 'bonus_leave_type id';
comment on column user_leave_requests.created_at is 'Save timestamp when create';
comment on column user_leave_requests.updated_at is 'Save timestamp when update';
comment on column user_leave_requests.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_leave_requests.organization_id is 'Id of organization';
comment on column user_leave_requests.user_id is 'Id of user';
comment on column user_leave_requests.leave_request_type_id is 'Id of leave request type';
comment on column user_leave_requests.datetime_leave_from is 'Datetime leave from';
comment on column user_leave_requests.datetime_leave_to is 'Datetime leave to';
comment on column user_leave_requests.created_by is 'Created by user';
comment on column user_leave_requests.updated_by is 'Updated by user';
comment on column user_leave_requests.email_title is 'Title of email';
comment on column user_leave_requests.email_content is 'Content of email';
comment on column user_leave_requests.reason is 'Reason';
comment on column user_leave_requests.hour is 'Hour leave caculate by date and time column. negative number';
