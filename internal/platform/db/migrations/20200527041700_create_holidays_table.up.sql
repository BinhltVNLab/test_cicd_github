create table if not exists holidays(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    organization_id integer not null,
    holiday_date date not null,
    description varchar (100)
);

create index index_holiday_holiday_date on holidays (holiday_date);

alter table holidays add constraint holiday_organization_id foreign key (organization_id) references organizations (id);

comment on column holidays.id is 'holidays id';
comment on column holidays.created_at is 'Save timestamp when create';
comment on column holidays.updated_at is 'Save timestamp when update';
comment on column holidays.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column holidays.organization_id is 'Organization id';
comment on column holidays.holiday_date is 'Holiday date';
comment on column holidays.description is 'Description';
