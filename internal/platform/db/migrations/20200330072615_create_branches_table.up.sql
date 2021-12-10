create table if not exists branches(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar(50) not null,
    organization_id int not null
);
create index index_branch_name on branches (name);
create index index_branch_organization_id on branches (organization_id);

comment on column branches.id is 'branch id';
comment on column branches.created_at is 'Save timestamp when create';
comment on column branches.updated_at is 'Save timestamp when update';
comment on column branches.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column branches.name is 'Name of branch';
comment on column branches.organization_id is 'Organization id';

alter table branches add constraint branch_organization foreign key (organization_id) references organizations (id);
alter table user_profiles add constraint user_profile_branch foreign key (branch) references branches (id);
