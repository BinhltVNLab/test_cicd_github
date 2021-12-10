create table if not exists noti_requests(
     id serial primary key not null,
     status varchar(100) not null,
     created_at timestamp not null,
     updated_at timestamp not null,
     deleted_at timestamp
);

alter table notifications
    add column if not exists title varchar(250) default '' not null;

alter table notifications
    add  column if not exists noti_request_id integer default 0 not null;

create index index_notifications_noti_request_id on notifications (noti_request_id);

create table if not exists email_noti_requests(
    id serial primary key not null,
    noti_request_id integer not null ,
    organization_id integer not null,
    sender integer not null,
    to_user_ids integer [] not null ,
    subject varchar not null ,
    content varchar not null ,
    url  varchar,
    template varchar ,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_email_noti_requests_request_id on email_noti_requests (noti_request_id);




