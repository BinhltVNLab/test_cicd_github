create table if not exists common_targets(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    organization_id integer not null,
    name varchar(100) not null,
    weight decimal,
    description text,
    quarter integer,
    year integer
);

create index index_organization_id on common_targets (organization_id);

comment on column common_targets.id is 'common_target id';
comment on column common_targets.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column common_targets.name is 'common target name';
comment on column common_targets.organization_id is 'Organization id. Foreign key to table organizations.id';
comment on column common_targets.description is 'common target description';
comment on column common_targets.created_at is 'Save timestamp when create';
comment on column common_targets.updated_at is 'Save timestamp when update';
comment on column common_targets.quarter is 'quater of the year';
comment on column common_targets.year is 'year';
comment on column common_targets.weight is 'point of the common target';
