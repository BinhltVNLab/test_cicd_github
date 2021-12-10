create table if not exists asset_types(
    id serial primary key not null,
    organization_id int not null,
    name varchar (150),
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_asset_type_organization_id on asset_types (organization_id);

comment on column asset_types.id is 'id';
comment on column asset_types.organization_id is 'organization id';
comment on column asset_types.name is 'asset type name';
comment on column asset_types.created_at is 'Save timestamp when create';
comment on column asset_types.updated_at is 'Save timestamp when update';
comment on column asset_types.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table asset_types add constraint asset_type_organization_id foreign key (organization_id) references organizations (id);
