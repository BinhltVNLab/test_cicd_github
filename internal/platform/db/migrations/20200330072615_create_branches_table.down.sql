alter table user_profiles drop constraint if exists user_profile_branch;
alter table branches drop constraint if exists branch_organization;
drop table if exists branches;
