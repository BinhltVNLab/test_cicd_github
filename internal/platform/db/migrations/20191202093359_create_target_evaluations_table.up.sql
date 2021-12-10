create table if not exists target_evaluations(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    user_id integer not null,
    content jsonb,
    is_processing boolean not null,
    organization_id integer not null,
    status integer not null,
    quarter integer not null,
    year integer not null,
    updated_by varchar(30) not null
);
create index index_target_evaluations_user_id on target_evaluations (user_id);

comment on column target_evaluations.id is 'target_evaluation_form id of user';
comment on column target_evaluations.user_id is 'User id, foreign key to table users.id';
comment on column target_evaluations.content is 'Target content';
comment on column target_evaluations.is_processing is 'check it is processing';
comment on column target_evaluations.status is 'Status of target_evaluation_form';
comment on column target_evaluations.quarter is 'Quarter when creating the target_evaluation_form';
comment on column target_evaluations.year is 'Year when creating the target_evaluation_form';
comment on column target_evaluations.updated_by is 'Updated by someone';
comment on column target_evaluations.created_at is 'Timestamp when created';
comment on column target_evaluations.updated_at is 'Timestamp when updated';
comment on column target_evaluations.deleted_at is 'Timestamp when deleted';
