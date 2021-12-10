create table if not exists modules(
    id serial primary key not null,
    name varchar (100) not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

comment on column modules.id is 'module id';
comment on column modules.created_at is 'Save timestamp when create';
comment on column modules.updated_at is 'Save timestamp when update';
comment on column modules.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column modules.name is 'module name';
