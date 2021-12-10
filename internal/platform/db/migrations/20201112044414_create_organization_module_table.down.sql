alter table organization_modules drop constraint if exists org_module_module_id;
alter table organization_modules drop constraint if exists org_module_organization_id;
drop table if exists organization_modules;
