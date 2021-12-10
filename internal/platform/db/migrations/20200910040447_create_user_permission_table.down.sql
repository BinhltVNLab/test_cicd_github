alter table user_profiles drop constraint if exists user_profile_permission;
alter table user_permission drop constraint if exists user_permission_organization;
drop table if exists user_permission;
