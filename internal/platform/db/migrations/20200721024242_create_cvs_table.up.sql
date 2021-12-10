create table if not exists cvs(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    recruitment_id integer not null,
    media_id integer not null,
    file_name varchar (255) not null,
    status smallint not null
);

create index index_cvs_recruitment_id on cvs (recruitment_id);
create index index_cvs_media_id on cvs (media_id);

alter table cvs add constraint fk_cvs_recruitment_id foreign key (recruitment_id) references recruitments (id);

comment on column cvs.id is 'cv id';
comment on column cvs.created_at is 'Save timestamp when create';
comment on column cvs.updated_at is 'Save timestamp when update';
comment on column cvs.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column cvs.recruitment_id is 'Id of recruitment';
comment on column cvs.media_id is 'Id of media';
comment on column cvs.file_name is 'File name';
comment on column cvs.status is 'Status';
