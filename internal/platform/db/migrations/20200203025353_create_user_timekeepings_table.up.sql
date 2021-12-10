create table if not exists user_timekeepings(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    organization_id integer not null,
    user_id integer not null,
    check_in_time timestamp,
    check_out_time timestamp
);

comment on column user_timekeepings.id is 'timekeeping id';
comment on column user_timekeepings.created_at is 'Save timestamp when create';
comment on column user_timekeepings.updated_at is 'Save timestamp when update';
comment on column user_timekeepings.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_timekeepings.user_id is 'Foreign key to users table';
comment on column user_timekeepings.organization_id is 'Foreign key to organizations table';
comment on column user_timekeepings.check_in_time is 'Check in time';
comment on column user_timekeepings.check_out_time is 'Checkout time';
