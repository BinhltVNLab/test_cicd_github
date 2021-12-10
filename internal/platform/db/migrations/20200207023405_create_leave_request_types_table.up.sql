create table if not exists leave_request_types(
    id int primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar(50) not null,
    description varchar(100) not null
);

create index index_leave_request_types_id on leave_request_types (id);

comment on column leave_request_types.id is 'leave_request_type id';
comment on column leave_request_types.created_at is 'Save timestamp when create';
comment on column leave_request_types.updated_at is 'Save timestamp when update';
comment on column leave_request_types.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column leave_request_types.name is 'Name of leave type';
comment on column leave_request_types.description is 'Description of leave type';
