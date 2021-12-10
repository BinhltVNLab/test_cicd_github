create table if not exists user_permissions(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    organization_id int not null,
    user_id int not null,
    modules jsonb
);
create index index_user_permission_organization_id on user_permissions (organization_id);

comment on column user_permissions.id is 'user permisstion id';
comment on column user_permissions.created_at is 'Save timestamp when create';
comment on column user_permissions.updated_at is 'Save timestamp when update';
comment on column user_permissions.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_permissions.organization_id is 'Organization id';
comment on column user_permissions.modules is 'modules of system';


alter table user_permissions add constraint user_permission_organization foreign key (organization_id) references organizations (id);
alter table user_permissions add constraint user_profile_permission foreign key (user_id) references  user_profiles (user_id);

