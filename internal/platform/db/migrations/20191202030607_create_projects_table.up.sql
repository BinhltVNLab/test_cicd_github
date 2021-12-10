create table if not exists projects(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar(100) not null,
    description text,
    targets jsonb,
    organization_id integer not null
);

create index index_projects_organization_id on projects (organization_id);

comment on column projects.id is 'projects id';
comment on column projects.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column projects.name is 'Name';
comment on column projects.organization_id is 'Organization id. Foreign key to table organizations.id';
comment on column projects.description is 'Description';
comment on column projects.created_at is 'Save timestamp when create';
comment on column projects.updated_at is 'Save timestamp when update';
comment on column projects.targets is 'target of the project. Example: [{year: 2019, quarter: 4, target: Finish 80% project}]';
