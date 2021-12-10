alter table common_targets add constraint common_targets_organization_id foreign key (organization_id) references organizations (id);
