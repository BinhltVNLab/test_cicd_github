alter table user_profiles drop constraint if exists user_profile_job_title;
alter table job_titles drop constraint if exists job_title_organization;
drop table if exists job_titles;
