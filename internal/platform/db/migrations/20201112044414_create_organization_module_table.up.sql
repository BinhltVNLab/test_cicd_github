create table if not exists organization_modules(
    id serial primary key not null,
    module_id int not null,
    organization_id int not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_org_module_module_id on organization_modules (module_id);
create index index_org_module_organization_id on organization_modules (organization_id);

comment on column organization_modules.id is 'id';
comment on column organization_modules.module_id is 'module id';
comment on column organization_modules.organization_id is 'organization id';
comment on column organization_modules.created_at is 'Save timestamp when create';
comment on column organization_modules.updated_at is 'Save timestamp when update';
comment on column organization_modules.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table organization_modules add constraint org_module_module_id foreign key (module_id) references modules (id);
alter table organization_modules add constraint org_module_organization_id foreign key (organization_id) references organizations (id);
