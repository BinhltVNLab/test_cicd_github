create table if not exists job_titles(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar(50) not null,
    organization_id int not null
);
create index index_job_title_name on job_titles (name);
create index index_job_title_organization_id on job_titles (organization_id);

comment on column job_titles.id is 'job_title id';
comment on column job_titles.created_at is 'Save timestamp when create';
comment on column job_titles.updated_at is 'Save timestamp when update';
comment on column job_titles.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column job_titles.name is 'Name of job title';
comment on column job_titles.organization_id is 'Organization id';

alter table job_titles add constraint job_title_organization foreign key (organization_id) references organizations (id);
alter table user_profiles add constraint user_profile_job_title foreign key (job_title) references job_titles (id);
