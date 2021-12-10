create table if not exists leave_bonus_types(
    id int primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar(50) not null,
    description varchar(100) not null
);

create index index_leave_bonus_types_id on leave_bonus_types (id);

comment on column leave_bonus_types.id is 'bonus_leave_type id';
comment on column leave_bonus_types.created_at is 'Save timestamp when create';
comment on column leave_bonus_types.updated_at is 'Save timestamp when update';
comment on column leave_bonus_types.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column leave_bonus_types.name is 'Name of bonus leave type';
comment on column leave_bonus_types.description is 'Description of bonus leave type';
