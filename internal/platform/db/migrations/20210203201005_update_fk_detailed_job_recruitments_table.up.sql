alter table detailed_job_recruitments drop constraint if exists detailed_job_recruitments_recruitment_id;

alter table detailed_job_recruitments
    add constraint detailed_job_recruitments_recruitment_id foreign key (recruitment_id)
        references recruitments on delete cascade;

