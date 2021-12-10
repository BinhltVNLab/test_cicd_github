create table if not exists cv_comments(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    cv_id integer not null,
    created_by integer not null,
    comment varchar (511) not null
);

create index index_cv_comments_cv_id on cv_comments (cv_id);
create index index_cv_comments_created_by on cv_comments (created_by);

alter table cv_comments add constraint fk_cv_comments_cv_id foreign key (cv_id) references cvs (id);
alter table cv_comments add constraint fk_cv_comments_created_by foreign key (created_by) references users (id);

comment on column cv_comments.id is 'cv id';
comment on column cv_comments.created_at is 'Save timestamp when create';
comment on column cv_comments.updated_at is 'Save timestamp when update';
comment on column cv_comments.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column cv_comments.cv_id is 'Id of cv';
comment on column cv_comments.created_by is 'Id of user';
comment on column cv_comments.comment is 'Content of comment';
