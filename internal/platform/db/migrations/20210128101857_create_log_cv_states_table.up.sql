create table if not exists log_cv_states(
    id serial primary key not null,
    cv_id int not null,
    status smallint not null,
    update_day date,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_log_cv_states_cv_id on log_cv_states (cv_id);

comment on column log_cv_states.id is 'id';
comment on column log_cv_states.cv_id is 'cv id';
comment on column log_cv_states.status is 'status';
comment on column log_cv_states.update_day is 'update day';
comment on column log_cv_states.created_at is 'Save timestamp when create';
comment on column log_cv_states.updated_at is 'Save timestamp when update';
comment on column log_cv_states.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table log_cv_states add constraint log_cv_states_cv_id foreign key (cv_id) references cvs (id);
