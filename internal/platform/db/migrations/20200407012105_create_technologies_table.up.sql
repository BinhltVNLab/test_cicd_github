create table if not exists technologies(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar(50) not null,
    organization_id int not null
);
create index index_technologies_name on technologies (name);
create index index_technologies_organization_id on technologies (organization_id);

comment on column technologies.id is 'technology id';
comment on column technologies.created_at is 'Save timestamp when create';
comment on column technologies.updated_at is 'Save timestamp when update';
comment on column technologies.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column technologies.name is 'Name of technology';
comment on column technologies.organization_id is 'Organization id';
