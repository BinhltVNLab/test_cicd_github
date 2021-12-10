create table if not exists contract_types(
    id serial primary key not null,
    organization_id int not null,
    name varchar (150) not null,
    file_template_name varchar (500) not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_contract_types_organization_id on contract_types (organization_id);

comment on column contract_types.id is 'id';
comment on column contract_types.organization_id is 'organization id';
comment on column contract_types.name is 'contract type name';
comment on column contract_types.file_template_name is 'name of file template';
comment on column contract_types.created_at is 'Save timestamp when create';
comment on column contract_types.updated_at is 'Save timestamp when update';
comment on column contract_types.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table contract_types add constraint contract_types_organization_id foreign key (organization_id) references organizations (id);
