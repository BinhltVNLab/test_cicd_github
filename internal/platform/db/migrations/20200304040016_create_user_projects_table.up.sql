create table if not exists user_projects(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    user_id int not null,
    project_id int not null
);
create index index_project on user_projects (project_id);
create index index_user_project on user_projects (project_id, user_id);

comment on column user_projects.id is 'user_project id';
comment on column user_projects.created_at is 'Save timestamp when create';
comment on column user_projects.updated_at is 'Save timestamp when update';
comment on column user_projects.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_projects.user_id is 'Id of user';
comment on column user_projects.project_id is 'Id of project';
