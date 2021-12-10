create table if not exists functions(
    id serial primary key not null,
    name varchar (100) not null,
    module_id int not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_function_module_id on functions (module_id);
create index index_function_name on functions (name);

comment on column functions.id is 'id';
comment on column functions.module_id is 'module id';
comment on column functions.name is 'function name';
comment on column functions.created_at is 'Save timestamp when create';
comment on column functions.updated_at is 'Save timestamp when update';
comment on column functions.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table functions add constraint function_module_id foreign key (module_id) references modules (id);
