create table if not exists notifications(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    organization_id integer not null,
    sender integer not null,
    receiver integer not null,
    content varchar (255),
    datetime_seen timestamp,
    redirect_url varchar (255),
    status smallint not null
);

create index index_notification_status on notifications (status);

alter table notifications add constraint fk_notifications_sender foreign key (sender) references users (id);
alter table notifications add constraint fk_notifications_receiver foreign key (receiver) references users (id);
alter table notifications add constraint fk_notifications_organization foreign key (organization_id) references organizations (id);

comment on column notifications.id is 'bonus_leave_type id';
comment on column notifications.created_at is 'Save timestamp when create';
comment on column notifications.updated_at is 'Save timestamp when update';
comment on column notifications.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column notifications.organization_id is 'Id of organization';
comment on column notifications.sender is 'Id of user send';
comment on column notifications.receiver is 'Id of user receive';
comment on column notifications.content is 'Content of id';
comment on column notifications.datetime_seen is 'Datetime seen notification';
comment on column notifications.redirect_url is 'Redirect url';
comment on column notifications.status is 'Status';
