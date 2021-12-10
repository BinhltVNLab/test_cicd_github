create table if not exists detailed_job_recruitments (
    id serial primary key not null,
    recruitment_id int not null,
    amount smallint not null,
    address VARCHAR (250)[] not null,
    role smallint not null,
    gender smallint,
    type_of_work VARCHAR (150) not null,
    experience smallint not null,
    salary_type smallint not null,
    salary_from bigint,
    salary_to bigint,
    profile_recipients VARCHAR (150),
    email VARCHAR(150),
    phone_number VARCHAR (50),
    description text,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_detailed_job_recruitments_recruitment_id on detailed_job_recruitments (recruitment_id);

comment on column detailed_job_recruitments.id is 'id';
comment on column detailed_job_recruitments.amount is 'Number of candidates';
comment on column detailed_job_recruitments.address is 'Company addresses';
comment on column detailed_job_recruitments.role is 'Role applying';
comment on column detailed_job_recruitments.gender is 'gender';
comment on column detailed_job_recruitments.type_of_work is 'Type of work: CTV, etc...';
comment on column detailed_job_recruitments.experience is 'Experience';
comment on column detailed_job_recruitments.salary_type is 'Salary type';
comment on column detailed_job_recruitments.salary_from is 'Salary from';
comment on column detailed_job_recruitments.salary_to is 'Salary to';
comment on column detailed_job_recruitments.profile_recipients is 'Profile recipients';
comment on column detailed_job_recruitments.email is 'Email';
comment on column detailed_job_recruitments.phone_number is 'Contact phone number';
comment on column detailed_job_recruitments.description is 'Description';

alter table detailed_job_recruitments add constraint detailed_job_recruitments_recruitment_id foreign key (recruitment_id) references recruitments (id);
