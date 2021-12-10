alter table projects add column managed_by INTEGER;
alter table user_projects add constraint user_projects_user_id foreign key (user_id) references users (id);
alter table user_projects add constraint user_projects_project_id foreign key (project_id) references projects (id);
