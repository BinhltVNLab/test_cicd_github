alter table projects add constraint projects_organization_id foreign key (organization_id) references organizations (id);
