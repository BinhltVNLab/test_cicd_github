create table if not exists recruitments(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    organization_id integer not null,
    job_name varchar (255) not null,
    description varchar (255) not null,
    start_date timestamp not null,
    expiry_date timestamp not null,
    branch_ids integer [] not null,
    assignees integer [] not null
);

create index index_recruitments_organization_id on recruitments (organization_id);
create index index_recruitments_job_name on recruitments (job_name);
create index index_recruitments_start_date on recruitments (start_date);
create index index_recruitments_expiry_date on recruitments (expiry_date);

alter table recruitments add constraint fk_recruitments_organization_id foreign key (organization_id) references organizations (id);

comment on column recruitments.id is 'recruitment id';
comment on column recruitments.created_at is 'Save timestamp when create';
comment on column recruitments.updated_at is 'Save timestamp when update';
comment on column recruitments.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column recruitments.organization_id is 'Id of organization';
comment on column recruitments.job_name is 'Job name';
comment on column recruitments.description is 'Job description';
comment on column recruitments.start_date is 'Start date';
comment on column recruitments.expiry_date is 'Expiry date';
comment on column recruitments.branch_ids is 'Id of branches';
comment on column recruitments.assignees is 'Id of assignee';
